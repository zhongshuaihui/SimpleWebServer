package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Book struct {
	Id        int    `json:"id"`
	Book_name string `json:"book_name" form:"book_name"`
	Author    string `json:"author" form:"author"`
	Price     int    `json:"price" form:"price"`
}

func main() {
	var err error
	//打开mysql数据库,parseTime表示解析时间
	db, err = sql.Open("mysql", "root:12345678@tcp(127.0.0.1:3306)/test?parseTime=true")
	if err != nil {
		log.Fatal(err.Error()) //输出错误，Fatal和panic的区别在前者不会执行defer
	}

	defer db.Close()
	// Open()并不会建立新的连接，需要Ping()来建立连接
	err = db.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}

	// 新建一个路由Handle
	router := gin.Default()

	// 返回所有书
	router.GET("/library", func(c *gin.Context) {
		// c.String(http.StatusOK, "Hello World!")
		var book Book
		books, err := book.getLibrary()
		if err != nil {
			log.Fatal(err)
		}
		// c.Json()创建json类型，gin.H将数据转换成map
		c.JSON(http.StatusOK, gin.H{"results": books, "count": len(books)})
	})

	// 返回某个ID的书，如/library/1
	router.GET("/library/:id", func(c *gin.Context) {
		// c.Param()可以得到url中的字段
		id_str := c.Param("id")
		id, err := strconv.Atoi(id_str)
		if err != nil {
			log.Fatal(err)
		}

		book := Book{
			Id: id,
		}
		this_book, err := book.getById()
		var result gin.H
		if err != nil {
			result = gin.H{"result": nil, "count": 0}
		} else {
			result = gin.H{"result": this_book, "count": 1}
		}
		c.JSON(http.StatusOK, result)
	})

	// 新插入一本书
	router.POST("/library", func(c *gin.Context) {
		var book Book
		// 获取要插入的书
		err := c.Bind(&book)
		if err != nil {
			log.Fatal(err)
		}
		// book_name := c.PostForm("book_name")
		// author := c.PostForm("author")
		// price_str := c.PostForm("price")
		// fmt.Println(book_name, author, price_str)
		// price, err := strconv.Atoi(price_str)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// book := Book{
		// 	Book_name: book_name,
		// 	Author:    author,
		// 	Price:     price,
		// }

		Id, err := book.addBook()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("ID = ", Id)
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s 插入成功", book.Book_name)})
	})

	router.DELETE("/library/:id", func(c *gin.Context) {
		id_str := c.Param("id")
		id, err := strconv.Atoi(id_str)
		if err != nil {
			log.Fatal(err)
		}

		book := Book{
			Id: id,
		}

		num := book.delBook()
		var result gin.H
		if num == 0 {
			result = gin.H{"result": "No book is deleted"}
		} else if num == 1 {
			result = gin.H{"result": "One kind of books are deleted"}
		} else {
			result = gin.H{"result": fmt.Sprintf("%d kind of books are deleted", num)}
		}
		c.JSON(http.StatusOK, result)
	})

	router.Run(":8080")
}

func (b Book) getLibrary() (books []Book, err error) {
	// 使用db.Query查询数据库
	rows, err := db.Query("select * from library")
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() { // 使用rows.Next()遍历返回的数据库所有行，当读到EOF时会自动推出
		var book Book
		rows.Scan(&book.Id, &book.Book_name, &book.Author, &book.Price) // 使用rows.Scan()读取所有行的数据
		books = append(books, book)
	}
	return
}

func (b Book) getById() (this_book Book, err error) {
	// QueryRow只返回一行查询，且不需要Next()操作
	row := db.QueryRow("select * from library where id = ?", b.Id)

	err = row.Scan(&this_book.Id, &this_book.Book_name, &this_book.Author, &this_book.Price)
	return
}

func (b Book) addBook() (Id int, err error) {
	// Prepare指令可先声明带占位符的语句，然后再将占位符替换
	stmt, err := db.Prepare("insert into library(book_name, author, price) values (?, ?, ?)")
	if err != nil {
		return
	}
	defer stmt.Close()

	// Exec执行一个query但不返回任何rows，它的返回值Result为一个特殊的interface
	rs, err := stmt.Exec(b.Book_name, b.Author, b.Price)
	if err != nil {
		return
	}

	// LastInsertId()为Result的一个(int64, error)返回
	id, err := rs.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	Id = int(id)
	return
}

func (b Book) delBook() (num int) {
	stmt, err := db.Prepare("delete from library where id = ?")
	if err != nil {
		return
	}
	defer stmt.Close()

	rs, err := stmt.Exec(b.Id)
	if err != nil {
		return
	}

	Num, err := rs.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	num = int(Num)
	return
}
