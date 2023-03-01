# SIDE technical test

## Overview

this repository hold the entry technical test for SIDE.

## build the repo

```sh
docker build -t side:1 ./
```

## run the repo

```sh
docker run -i -t -p8888:80 -v $PWD/side.db:/code/side.db side:1
```

This will allow you to open the sqlite3 database on your host to inspect the content if necessary.
