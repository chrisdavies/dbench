# DBENCH

Basic benchmarks for SQLite vs filesystem (btrfs on a 2020 Dell XPS SSD).

## Linear writes

10k inserts, written in a tight loop.

```
SQLite      647ms, 809ms, 708ms
FileSystem  393ms, 371ms, 382ms
```

For sequential writes, the file system is 1.8x faster.

## Concurrent writes

Here, we're going to crank up the concurrency and have 100 concurrent writers, each writing 1K records. SQLite will serialize these under the hood. The filesystem will do whatever it does.

```
SQLite        8.2s, 8.1s, 8.5s
FileSystem    3.1s, 3.4s, 3.4s
```

For concurrent writes, the file system is 2.4x faster.

This is not terribly surprising, as SQLite does not handle concurrent writes. To be honest the performance gap here is smaller than I'd have guessed.

- SQLite can write around 12K inserts per second
- BTRFS can write around 30K files per second

## A more realistic test

The real-world application I'll be writing would have a bit more structure and more indices. So, I think the next test will be to run N tasks from start to finish: insert, update status, update progress, delete.

For this, we'll want to index by status, and in the real world, I'd *probably* also index by `scheduled_at` so that we can handle job scheduling efficiently. If I were to use the file system for this, I'd keep the queues and schedules in memory, and rebuild it when the application starts. I've tested the in-memory approach, and it's blisteringly fast (millions of ops per second, given proper care).


