package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server celler server.
// @termsOfService http://swagger.io/terms/
func main() {
	fmt.Println("Simple Web Server")

	r := mux.NewRouter()
	r.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
	r.HandleFunc("/orders", handleOrder)

	http.ListenAndServe(":8181", r)
}

func handleOrder(res http.ResponseWriter, req *http.Request) {

	//Req log check
	fmt.Println("message received from ", req.Host, req.URL.Path, req.URL.Query(), req.Method, req.Body)

	//Connect To DB
	db, err := connectDB()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("db conn ok")
	defer db.Close()

	//if method == POST then inseert into db
	if req.Method == "POST" {

		//Deserailize json body to each obj
		data := &Order{}
		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		insert1, err := db.Prepare("insert into orders(customer_name) values (?)")
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		tmp, _ := insert1.Exec(data.CustomerName)

		fmt.Println(tmp.LastInsertId())
		order_id, _ := tmp.LastInsertId()
		insert2, err := db.Query("insert into items(item_code, description, quantity, order_id) values (?, ?, ?, ?)", data.Items.ItemCode, data.Items.Description, data.Items.Quantity, order_id)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		insert1.Close()
		insert2.Close()

		res.Write([]byte("Insert Success"))

	}

	//If method == get then return all orders
	if req.Method == "GET" {

		var orderList []Order
		rows, err := db.Query(
			"select customer_name, ordered_at, item_code, description, quantity from orders inner join items on orders.order_id = items.order_id")

		defer rows.Close()

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for rows.Next() {
			var each = Order{}
			var err = rows.Scan(&each.CustomerName, &each.OrderedAt, &each.Items.ItemCode, &each.Items.Description, &each.Items.Quantity)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			orderList = append(orderList, each)
		}

		json, _ := json.Marshal(orderList)
		fmt.Println(string(json))

		res.Write(json)

	}

	if req.Method == "PUT" {
		//Deserailize json body to each obj
		data := &Order{}
		err := json.NewDecoder(req.Body).Decode(&data)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Println(data)

		update1, err := db.Query("UPDATE orders SET customer_name = ?, ordered_at = ? where order_id = ?", data.CustomerName, data.OrderedAt, data.OrderID)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
		update2, err := db.Query("UPDATE items SET item_code = ?, description = ?, quantity = ? where items.order_id = ?", data.Items.ItemCode, data.Items.Description, data.Items.Quantity, data.OrderID)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		update1.Close()
		update2.Close()

		res.Write([]byte("Update Success"))

	}

	if req.Method == "DELETE" {
		data := req.URL.Query()
		delete1, err := db.Query(
			"DELETE FROM orders WHERE order_id = ?", data.Get("id"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		delete2, err := db.Query("DELETE FROM items WHERE order_id = ?", data.Get("id"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		delete1.Close()
		delete2.Close()

		res.Write([]byte("Delete Success"))
	}

	return
}

func connectDB() (*sql.DB, error) {
	//DSN following my mysql config
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3308)/order_by?parseTime=true")

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return db, nil
}

//Struct
type Order struct {
	OrderID      int       `json:"orderId"`
	OrderedAt    time.Time `json:"orderedAt"`
	CustomerName string    `json:"customerName"`
	Items        struct {
		ItemID      int    `json:"itemID"`
		ItemCode    string `json:"itemCode"`
		Description string `json:"description"`
		Quantity    int    `json:"quantity"`
	} `json:"items"`
}
