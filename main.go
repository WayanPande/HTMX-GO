package main

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type book struct {
	ID       int    `json:"id"`
	Title    string `json:"title" form:"title"`
	Author   string `json:"author" form:"author"`
	Quantity int    `json:"quantity" form:"quantity"`
}

var books = []book{
	{ID: 1, Title: "In Search of Lost Time", Author: "Marcel Proust", Quantity: 2},
	{ID: 2, Title: "The Greate Gatsbye", Author: "Scott Fitzgerald", Quantity: 5},
	{ID: 3, Title: "War and Pieece", Author: "Leo Tolstoy", Quantity: 7},
}

func getBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, books)
}

func getBookById(id int) (*book, error) {
	for i, b := range books {
		if b.ID == id {
			return &books[i], nil
		}
	}
	return nil, errors.New("book not found")
}

func bookById(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	book, err := getBookById(i)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, book)
}

func createBooks(c *gin.Context) {
	var newBook book

	if err := c.BindJSON(&newBook); err != nil {
		return
	}

	books = append(books, newBook)
	c.IndentedJSON(http.StatusCreated, newBook)
}

func checkoutBook(c *gin.Context) {
	id, ok := c.GetQuery("id")
	i, err := strconv.Atoi(id)
	if err != nil {

		panic(err)
	}

	if !ok {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Missing id query params"})
		return
	}

	book, err := getBookById(i)

	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found"})
		return
	}

	if book.Quantity <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Book not available"})
		return
	}

	book.Quantity -= 1
	c.IndentedJSON(http.StatusOK, book)
}

func GetData(c context.Context) map[string][]book {
	db := GetConnection()

	script := "SELECT * FROM book"
	rows, err := db.QueryContext(c, script)
	db.Close()
	if err != nil {
		panic(err)
	}

	var books []book

	for rows.Next() {
		temp_book := book{}
		rows.Scan(&temp_book.ID, &temp_book.Title, &temp_book.Author, &temp_book.Quantity)
		books = append(books, temp_book)
	}
	rows.Close()

	booksReturn := map[string][]book{
		"Books": books,
	}

	return booksReturn
}

func renderPage(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", GetData(c))
}

func addBook(c *gin.Context) {
	var person book

	if c.ShouldBind(&person) == nil {
		// log.Println(person.Title)
		// log.Println(person.Author)
		// log.Println(person.Quantity)

		db := GetConnection()
		script := "INSERT INTO book(title, author, quantity) VALUES (?, ?, ?)"
		data, err := db.ExecContext(c, script, person.Title, person.Author, person.Quantity)
		if err != nil {
			panic(err)
		}
		id, err := data.LastInsertId()
		if err != nil {
			panic(err)
		}

		db.Close()
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		tmpl.ExecuteTemplate(c.Writer, "film-list-element", book{Title: person.Title, Author: person.Author, Quantity: person.Quantity, ID: int(id)})
	}
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")

	db := GetConnection()
	script := "DELETE FROM book WHERE id=(?)"
	_, err := db.ExecContext(c, script, id)
	if err != nil {
		panic(err)
	}

	db.Close()
}

func main() {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/", renderPage)

	router.POST("/add-book", addBook)

	router.DELETE("/book/:id", deleteBook)

	// Rest API Stuff
	router.GET("/books", getBooks)
	router.POST("/books", createBooks)
	router.GET("/books/:id", bookById)
	router.PATCH("/checkout", checkoutBook)
	router.Run(":42069")
}
