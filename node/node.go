package node

import (
	"encoding/json"
	"fmt"
	"github.com/kparkins/yarbit/database"
	"io/ioutil"
	"net/http"
)

const Port = 80

type ErrorResponse struct {
	Error string `json:"error"`
}

type BalancesListResponse struct {
	Hash     database.Hash             `json:"block_hash"`
	Balances map[database.Account]uint `json:"balances"`
}

type TxAddRequest struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

type TxAddResponse struct {
	Hash database.Hash `json:"block_hash"`
}

func Run(dataDir string) error {
	state, err := database.NewStateFromDisk(dataDir)
	if err != nil {
		return err
	}
	defer state.Close()

	http.HandleFunc("/balances/list", func(writer http.ResponseWriter, request *http.Request) {
		balanceListHandler(writer, request, state)
	})
	http.HandleFunc("/tx/add", func(writer http.ResponseWriter, request *http.Request) {
		txAddHandler(writer, request, state)
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
}

func balanceListHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	writeResponse(w, BalancesListResponse{Hash: state.LatestBlockHash(), Balances: state.Balances})
}

func txAddHandler(w http.ResponseWriter, r *http.Request, state *database.State) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	txRequest := TxAddRequest{}
	if err := json.Unmarshal(content, &txRequest); err != nil {
		writeErrorResponse(w, err, http.StatusBadRequest)
		return
	}
	tx := database.NewTx(
		database.NewAccount(txRequest.From),
		database.NewAccount(txRequest.To),
		txRequest.Value,
		txRequest.Data,
	)
	if err := state.AddTx(tx); err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	hash, err := state.Persist()
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	writeResponse(w, TxAddResponse{Hash: hash})
}

func writeResponse(w http.ResponseWriter, content interface{}) {
	contentJson, err := json.Marshal(content)
	if err != nil {
		writeErrorResponse(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(contentJson)
}

func writeErrorResponse(w http.ResponseWriter, e error, statusCode int) {
	w.WriteHeader(statusCode)
	if e != nil {
		contentJson, _ := json.Marshal(ErrorResponse{Error: e.Error()})
		w.Header().Set("Content-Type", "application/json")
		w.Write(contentJson)
	}
}
