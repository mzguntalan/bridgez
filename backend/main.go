package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func handler(w http.ResponseWriter, r *http.Request){
    fmt.Fprintf(w, "Hello %s", r.URL.Path[1:])
}

func insertWordCmd(word string) (sqlCommand string) {
    return fmt.Sprintf("insert into words (representation) values ('%s') on conflict (representation) do nothing;", word )
}

func insertWordToDB(dbpool *pgxpool.Pool, word string) (err error) {
    _, err = dbpool.Exec(context.Background(), insertWordCmd(word)) 
    return err
} 

func insertWordsToDB(dbpool *pgxpool.Pool, words []string) (err error) {
    batch := &pgx.Batch{ }
    for _, word := range words {
        batch.Queue(insertWordCmd(word))
    }

    batchCmd := dbpool.SendBatch(context.Background(), batch)
    _, err =  batchCmd.Exec()
    return err
}


func insertLinkCmd(sourceWordID int, targetWordID int) (sqlCommand string) {
    return fmt.Sprintf("insert into links (source_word_id, target_word_id) values ('%s', '%s') on conflict (source_word_id, target_word_id) do nothing;", strconv.Itoa(sourceWordID), strconv.Itoa(targetWordID))
}

func insertLinkToDB(dbpool *pgxpool.Pool, sourceWordID int, targetWordID int) (err error) {
    _, err = dbpool.Exec(context.Background(), insertLinkCmd(sourceWordID, targetWordID))
    return err
}

type Link struct {
    sourceWordID int
    targetWordID int
}

func insertLinksToDB(dbpool *pgxpool.Pool, links []Link) (err error) {
    batch := &pgx.Batch{ }
    for _, link := range links {
        batch.Queue(insertLinkCmd(link.sourceWordID, link.targetWordID))
    } 
    batchCmd := dbpool.SendBatch(context.Background(), batch)
    _, err = batchCmd.Exec()

    return err
}

func findWordInDB(dbpool *pgxpool.Pool, wordRepresentation string) (wordId int) {
    sqlCmd := fmt.Sprintf("SELECT word_id FROM words WHERE representation = '%s';", wordRepresentation)

    err := dbpool.QueryRow(context.Background(), sqlCmd).Scan(&wordId)
    if err == pgx.ErrNoRows {
        return -1 
    } 

    return wordId 
}

type DuplicateWordErr struct { wordWithDuplicate string }
func (m *DuplicateWordErr) Error() string {
    return fmt.Sprintf("Duplicate of %s found", m.wordWithDuplicate)
}

func editWordInDB(dbpool *pgxpool.Pool, wordRepresentation string, newWordRepresentation string) (err error) {
    newWordID := findWordInDB(dbpool, newWordRepresentation)
    if newWordID != -1 {
        return &DuplicateWordErr{ wordWithDuplicate: newWordRepresentation }
    }
    sqlCmd := fmt.Sprintf("update words set representation = '%s' where representation = '%s'", newWordRepresentation, wordRepresentation)
    _, err = dbpool.Exec(context.Background(), sqlCmd)
    return err
}

type NodeReactForceGraph struct {
    Id string `json:"id"`
    Name string `json:"name"`
}
type LinkReactForceGraph struct {
    Source string `json:"source"`
    Target string `json:"target"`
}

func wordRowToNodeReactForceGraph(row pgx.CollectableRow) (result NodeReactForceGraph, err error) {
	var representation string
    err = row.Scan(&representation)
    if err != nil {
        log.Println("ERROR")
        return NodeReactForceGraph{Id:"", Name:""}, err
    }

    log.Println(representation)
    result = NodeReactForceGraph { Id: representation, Name: representation }

    log.Println("result")
    log.Println(result)

    return result, nil 


}
func getAllWordsAsNodeReactForceGraph(dbpool *pgxpool.Pool) (results []NodeReactForceGraph, err error) {
    rows, err := dbpool.Query(context.Background(), "select representation from words")
    if err != nil {
        return nil, err 
    }

    results, err = pgx.CollectRows(rows, wordRowToNodeReactForceGraph)
    if err != nil {
        return nil, err
    }

    log.Println("received", results)
    return results, nil
}


func enableCors(w *http.ResponseWriter) {
    (*w).Header().Set("Access-Control-Allow-Origin", "*")
    (*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
    (*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
}

func handleGetAllWords(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
    enableCors(&w)
    wordsAsNodes, err := getAllWordsAsNodeReactForceGraph(dbpool)
    if err != nil {
        panic("Error has occured!")
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(wordsAsNodes)

}


func getAllLinksAsLinkReactForceGraph(dbpool *pgxpool.Pool) (results []LinkReactForceGraph, err error) {
    rows, err := dbpool.Query(context.Background(), "select source_words.representation as source, target_words.representation as target from links l left join words source_words on source_words.word_id = l.source_word_id left join words target_words on target_words.word_id = l.target_word_id;")
    if err != nil {
        log.Println("ERROR", err)
        return nil, err
    }

    results, err = pgx.CollectRows(rows, pgx.RowToStructByPos[LinkReactForceGraph])
    
    if err != nil {
        log.Println(err)
        log.Println("error occured in query links")
        return nil, err
    }

    log.Println(">> Links", results)

    return results, nil
}


func handleGetAllLinks(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
    enableCors(&w)
    links, err := getAllLinksAsLinkReactForceGraph(dbpool)
    if err != nil {
        panic("Something happened and it shouldn't have happened.")
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(links)
}

func handlePostNode(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
    enableCors(&w)
    if r.Method != "POST" {
        return
    }

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        panic("panicked while reading node post")
    }

    var t NodeReactForceGraph
    err = json.Unmarshal(body, &t)
    log.Println(t.Id, t.Name)
    if err != nil {
        panic("panicked while unmarshal node post")
    }

    repr := t.Id
    err = insertWordToDB(dbpool, repr)
    if err != nil {
        panic("panic at inserting the word")
    }
    w.WriteHeader(http.StatusOK)
}


func handlePostLink(w http.ResponseWriter, r *http.Request, dbpool *pgxpool.Pool) {
    enableCors(&w)
    if r.Method != "POST" {
        return
    }

    body, err := io.ReadAll(r.Body)
    defer r.Body.Close()
    if err != nil {
        panic("panicked while reading link post")
    }

    var t LinkReactForceGraph
    err = json.Unmarshal(body, &t)
    log.Println("link", t.Source, t.Target)
    if err != nil {
        panic("panicked while unmarshal link post")
    }

    source := t.Source
    target := t.Target
    err = insertLinkToDB(dbpool, findWordInDB(dbpool, source), findWordInDB(dbpool, target)) 
    if err != nil {
        panic("panic at inserting the link")
    }
    w.WriteHeader(http.StatusOK)
}

func main() {
    dbpool, err := pgxpool.New(context.Background(), os.Getenv("GOOSE_DBSTRING"))
    fmt.Println("Starting")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'HELLO';").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(greeting)

    fmt.Print("will insert neko")
    err = insertWordToDB(dbpool, "猫")
    fmt.Println("done inserting neko")
    err = insertWordToDB(dbpool, "犬")
    fmt.Println("done inserting inu")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Insert of word failed %v\n", err)
        os.Exit(1)
    }
    fmt.Println("done inserting. proceeding to finding")


    nekoID := findWordInDB(dbpool, "猫")
    inuID := findWordInDB(dbpool, "犬")
    fmt.Printf("neko %d, inu %d \n", nekoID, inuID)
    err = insertLinkToDB(dbpool, inuID, nekoID);

    if err != nil {
        fmt.Fprintf(os.Stderr, "Insert of link failed %v\n", err)
        os.Exit(1)
    }

    err = insertLinkToDB(dbpool, inuID, nekoID);

    if err != nil {
        fmt.Fprintf(os.Stderr, "Insert of link failed %v\n", err)
        os.Exit(1)
    }

    err = insertWordToDB(dbpool, "女")
    err = insertWordToDB(dbpool, "花")
    // err = editWordInDB(dbpool, "女", "男")

    var customErr *DuplicateWordErr
    if errors.As(err, &customErr) {
        fmt.Println("error")
    }

    words := [...]string{"女", "子供"}
    err = insertWordsToDB(dbpool, words[:])
    if err != nil {
        fmt.Fprintf(os.Stderr, "bulk insert of words failed %v\n", err)
        os.Exit(1)
    }

    links := [...]Link{{ sourceWordID: 1, targetWordID: 2 }, { sourceWordID: 2, targetWordID: 1 }, { sourceWordID: 99, targetWordID: 100}}
    err = insertLinksToDB(dbpool, links[:])

    if err != nil {
        fmt.Fprintf(os.Stderr, "bulk insert of links failed %v\n", err)
        os.Exit(1)
    }


    http.HandleFunc("/", handler)
    http.HandleFunc("/words", func(w http.ResponseWriter, r *http.Request){handleGetAllWords(w, r, dbpool)})
    http.HandleFunc("/links", func(w http.ResponseWriter, r *http.Request){handleGetAllLinks(w, r, dbpool)})
    http.HandleFunc("/word/add", func(w http.ResponseWriter, r *http.Request){handlePostNode(w, r, dbpool)})
    http.HandleFunc("/link/add", func(w http.ResponseWriter, r *http.Request){handlePostLink(w, r, dbpool)})
    log.Fatal(http.ListenAndServe(":8080", nil))

}
