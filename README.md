`go build` will create the `infomodels` binary.

### How to load data

```
go build && go install &&  infomodels --model pedsnet-core --modelv 2.3.0 load -s nemours_pedsnet -d 'postgresql://localhost:5433/pedsnet_dcc_v23?sslmode=disable' ~/Documents/PEDSnet/testdata
```

A little bit fussy. If you run it twice in a row, it will abort since it won't be able to create tables the second time around.  There is no revert/undo feature yet.
