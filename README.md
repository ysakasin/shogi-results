# shogi-results

Scrape game results which are held by Japan Shogi Association.

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
./shogi-result -m 201804
```
