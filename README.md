# shogi-result

Scrape game results which are hold by Japan Shogi Association.

Scraped data are available at `/results`.

## Build

```sh
dep encure
go build
```

## Useage

### Scrape all results

```sh
./shogi-result
```

### Scrape monthly results

```sh
./shogi-result -d 201804
```
