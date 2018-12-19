package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	pb "github.com/alxego/discounter/proto/go"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func generateDB() map[int]float32 {
	dbSize := 20
	db := make(map[int]float32, dbSize)
	for i := 0; i < dbSize; i++ {
		db[i] = rand.Float32()
	}
	return db
}

type priceGetter struct {
	Client pb.PricerClient
	DB     map[int]float32
}

func (p priceGetter) getDiscount(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		handleError(w, err)
	}
	price, err := p.getPrice(id)
	if err != nil {
		handleError(w, err)
	}

	responseJSON := struct {
		Price      float32 `json:"price,omitempty"`
		Discount   float32 `json:"discount,omitempty"`
		FinalPrice float32 `json:"final_price,omitempty"`
	}{
		Price:      price.Price,
		Discount:   p.DB[int(id)],
		FinalPrice: ((1.0 - p.DB[int(id)]) * price.Price),
	}
	if err = json.NewEncoder(w).Encode(responseJSON); err != nil {
		log.Println(err)
	}
	return
}

func (p priceGetter) getPrice(id int64) (price *pb.ItemPrice, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	price, err = p.Client.GetPrice(ctx, &pb.ItemID{ID: id})
	return
}

func handleError(w http.ResponseWriter, err error) {
	log.Println(err)
	responseJSON := struct {
		Err string `json:"err,omitempty"`
	}{
		Err: err.Error(),
	}
	if err := json.NewEncoder(w).Encode(responseJSON); err != nil {
		log.Println(err)
	}
}

func main() {
	conn, err := grpc.Dial(":8787", grpc.WithInsecure())
	if err != nil {
		return
	}
	defer conn.Close()

	pricer := priceGetter{
		Client: pb.NewPricerClient(conn),
		DB:     generateDB(),
	}

	router := mux.NewRouter()
	router.HandleFunc("/discount", pricer.getDiscount).Queries("id", "{id}").Methods("GET")

	http.ListenAndServe(":8989", router)
	return
}
