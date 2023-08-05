# Go distributed Lock

POC on different locking mechanism.


1. etcd
2. [minio/dsync](https://blog.min.io/minio-dsync-a-distributed-locking-and-syncing-package-for-go/)
3. redislock
4. redis (custom lock) see [here](https://github.com/alextanhongpin/go-redis-function)
5. Postgres advisory lock, see [here](https://github.com/alextanhongpin/dbtx/tree/master/postgres/lock)


Postgres also has it's own locking mechanism like `select for update ... nowait` etc.
