package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gorilla/mux"
)

type BlockData struct {
	USDollar float64
}

type Block struct {
	Index int
	Data BlockData
	Timestamp string
	Hash string
	PrevHash string
}

type Blockchain struct {
	chains []Block
}

func (block *Block) calculateHash() {
	record := string(block.Index)+strconv.FormatFloat(block.Data.USDollar, 'f', 1, 64)+block.Timestamp+block.PrevHash
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	block.Hash = hex.EncodeToString(hashed)
}

func NewBlockchain() Blockchain {
	bc := Blockchain{}
	genesisBlock := Block{
		Index: 0,
		Data: BlockData{USDollar: 0},
		Timestamp: time.Now().String(),
		Hash: "",
		PrevHash: "",
	}
	bc.chains = append(bc.chains, genesisBlock)
	return bc;
}

func (bc *Blockchain) appendNewBlock(newBlock *Block) {
	bc.chains = append(bc.chains, *newBlock)
}

func (bc *Blockchain) GetLatestBlock() Block {
	return bc.chains[len(bc.chains)-1]
}

func (bc *Blockchain) WriteNewBlock(data BlockData) {
	newBlock := Block{
		Index: len(bc.chains),
		Data: data,
		Timestamp: time.Now().String(),
		Hash: "",
		PrevHash: bc.chains[len(bc.chains)-1].Hash,
	}
	newBlock.calculateHash()
	bc.appendNewBlock(&newBlock)
}

func (bc *Blockchain) FormatJSON() (string, error) {
	bytes, err := json.MarshalIndent(bc.chains, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

var blockchain Blockchain

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	res, _ := blockchain.FormatJSON()
	io.WriteString(w, res)
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var bcData BlockData

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&bcData); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()

	blockchain.WriteNewBlock(bcData)
	respondWithJSON(w, r, http.StatusCreated, blockchain.GetLatestBlock())
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func run() error {
    mux := makeMuxRouter()
    httpAddr := "8081"
    log.Println("Listening on ", httpAddr)
    s := &http.Server{
        Addr:           ":" + httpAddr,
        Handler:        mux,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    if err := s.ListenAndServe(); err != nil {
        return err
    }
    return nil
}

func main() {
	blockchain = NewBlockchain()
	log.Fatal(run())
}

