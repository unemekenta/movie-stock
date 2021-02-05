package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// DBConn PostgreSQLに接続
func DBConn() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("envファイルの読み込みに失敗しました。 \n", err)
	}
	dsn := fmt.Sprintf("%s://%s:%s/%s?sslmode=disable", os.Getenv("PSQL_USER"), os.Getenv("PSQL_PASS"), os.Getenv("PSQL_PORT"), os.Getenv("PSQL_DBNAME"))
	log.Print("PostgreSQL DBに接続しています...")
	db, err := gorm.Open("postgres", dsn)
	log.Println("接続しました。")

	if err != nil {
		panic(err)
	}
	// 赤文字のログを出す
	// db.LogMode(true)
	return db
}

// User テーブル名はusers
type User struct {
	ID        string `gorm:"primary_key"`
	Name      string
	CreatedAt time.Time
}

func getUserAll(w http.ResponseWriter, r *http.Request) {
	var users []User
	db := DBConn()
	defer db.Close()

	db.Order("id").Find(&users)
	res, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
	return
}

func getUser(w http.ResponseWriter, r *http.Request) {
	var user []User
	db := DBConn()
	defer db.Close()

	params := mux.Vars(r)
	db.Where("id = ?", params["id"]).First(&user)
	res, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
	return
}

func createUser(w http.ResponseWriter, r *http.Request) {
	user := User{}
	db := DBConn()
	defer db.Close()

	name := r.FormValue("name")
	user.Name = name
	db.Create(&user)
}

func editUser(w http.ResponseWriter, r *http.Request) {
	userBefore := User{}
	db := DBConn()
	defer db.Close()

	params := mux.Vars(r)
	id := params["id"]
	userBefore.ID = id
	name := r.FormValue("name")
	userAfter := userBefore
	db.First(&userAfter)
	userAfter.Name = name
	db.Model(&userBefore).Update(&userAfter)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	users := User{}
	db := DBConn()
	defer db.Close()

	params := mux.Vars(r)
	id := params["id"]
	users.ID = id
	db.First(&users)
	db.Delete(&users)
}

func main() {
	r := mux.NewRouter()

	// ルート(エンドポイント)
	r.HandleFunc("/api/users", getUserAll).Methods("GET")
	r.HandleFunc("/api/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/api/users", createUser).Methods("POST")
	r.HandleFunc("/api/users/edit/{id}", editUser).Methods("POST")
	r.HandleFunc("/api/users/delete/{id}", deleteUser).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", r))
}
