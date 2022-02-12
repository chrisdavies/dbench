# DBENCH

Basic benchmarks for SQLite vs filesystem (btrfs on a 2020 Dell XPS SSD).

## Linear writes

10k inserts, written in a tight loop.

```
SQLite      647ms, 809ms, 708ms
FileSystem  393ms, 371ms, 382ms
```

For sequential writes, SQLite is roughly 1/2 as fast. Not bad, really.

## Concurrent writes

Here, we're going to crank up the concurrency and have 100 concurrent writers. SQLite will serialize these under the hood. The filesystem will do whatever it does.


