# bridgez

graph-centric app for language learning

## Demo

![animation](https://github.com/mzguntalan/media_for_other_repo/blob/main/bridgez/bridgez_demo.gif?raw=true)

## Setup

### Prerequisites

- node, npm
- golang
- goose
- postgresql with a database name of your choice
- goose strings setup with your database

### Serving

0. Do the migrations by going to the `mgrations` dir and doing `goose up`
1. Serve the frontend app by going into the `bridgez` dir and do `npm install` then `npm run dev`
2. Serve the backend by going into the `backend` dir and running `go run main.go` and make sure it is o 8080

## Notes

A quick prototype of an app for myself for Japanese learning. It's main focus is on keep words that appear together, together in the graph. The arrows represet connections of any kind.
