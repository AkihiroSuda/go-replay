# Reproduce [etcd #5155](https://github.com/coreos/etcd/issues/5155) data race

## What is the issue?
`TestIssue2746` causes data race intermittently:
(copied from the original issue text)
```
WARNING: DATA RACE
Read by goroutine 171:
  github.com/coreos/etcd/raft.(*node).step()
      /Users/xiangli/go/src/github.com/coreos/etcd/raft/node.go:415 +0x7b
  github.com/coreos/etcd/raft.(*node).Propose()
      /Users/xiangli/go/src/github.com/coreos/etcd/raft/node.go:390 +0x152
  github.com/coreos/etcd/etcdserver.(*EtcdServer).sync.func1()
      /Users/xiangli/go/src/github.com/coreos/etcd/etcdserver/server.go:860 +0x98

Previous write by goroutine 7:
  github.com/coreos/etcd/raft.(*node).run()
      /Users/xiangli/go/src/github.com/coreos/etcd/raft/node.go:328 +0xf66

Goroutine 171 (running) created at:
  github.com/coreos/etcd/etcdserver.(*EtcdServer).sync()
      /Users/xiangli/go/src/github.com/coreos/etcd/etcdserver/server.go:862 +0x2a5
  github.com/coreos/etcd/etcdserver.(*raftNode).start.func1()
      /Users/xiangli/go/src/github.com/coreos/etcd/etcdserver/raft.go:229 +0x1296

Goroutine 7 (running) created at:
  github.com/coreos/etcd/raft.StartNode()
      /Users/xiangli/go/src/github.com/coreos/etcd/raft/node.go:203 +0x827
  github.com/coreos/etcd/etcdserver.startNode()
      /Users/xiangli/go/src/github.com/coreos/etcd/etcdserver/raft.go:315 +0xbb5
  github.com/coreos/etcd/etcdserver.NewServer()
      /Users/xiangli/go/src/github.com/coreos/etcd/etcdserver/server.go:320 +0x4318
  github.com/coreos/etcd/integration.(*member).Launch()
      /Users/xiangli/go/src/github.com/coreos/etcd/integration/cluster.go:543 +0x72
  github.com/coreos/etcd/integration.(*cluster).Launch.func1()
      /Users/xiangli/go/src/github.com/coreos/etcd/integration/cluster.go:141 +0x2e
```

We can reproduce the issue by just setting two GoReplay injection points:
[replay-etcd-5155.patch](replay-etcd-5155.patch)
```diff
 diff --git a/raft/node.go b/raft/node.go
index 727dec0..09da485 100644
--- a/raft/node.go
+++ b/raft/node.go
@@ -19,6 +19,8 @@ import (
 
 	pb "github.com/coreos/etcd/raft/raftpb"
 	"golang.org/x/net/context"
+
+	"github.com/AkihiroSuda/go-replay"
 )
 
 type SnapshotStatus int
@@ -325,6 +327,7 @@ func (n *node) run(r *raft) {
 				// block incoming proposal when local node is
 				// removed
 				if cc.NodeID == r.id {
+					replay.Inject([]byte("run"))
 					n.propc = nil
 				}
 				r.removeNode(cc.NodeID)
@@ -412,6 +415,7 @@ func (n *node) ProposeConfChange(ctx context.Context, cc pb.ConfChange) error {
 func (n *node) step(ctx context.Context, m pb.Message) error {
 	ch := n.recvc
 	if m.Type == pb.MsgProp {
+		replay.Inject([]byte("step"))
 		ch = n.propc
 	}
 
```


## How to reproduce

 * Apply `replay-etcd-5155.patch` to [etcd@d32113a0e](https://github.com/coreos/etcd/commit/d32113a0e).
 * Build `integration.test` (`go test -race -c github.com/coreos/etcd/integration`).
 * Set `GOMAXPROCS=1`, `GRMAX=1000ms`.
 * Run `./integration.test -test.v -test.run TestIssue2746` repeatedly with an arbitrary `GRSEED`. If you hit the issue, you should be able to replay the scenario with that `GRSEED`. (But not so deterministically replayable at the moment)


You can use `Dockerfile` for ease of testing:
```
$ docker build -t etcd-5155 -f ../../Dockerfile.example.etcd-5155 ../..
$ for f in $(seq 1 100);do echo === GRSEED=$f ===; docker run -it --rm -e GRSEED=$f etcd-5155; done
```

### Reproducibility
Tested 50 times for each, on Ubuntu 15.10 amd64, Xeon E3-1220 v3 (4 cores).
I used GoReplay@[95623afb](https://github.com/AkihiroSuda/go-replay/commit/95623afb).

Seed|Reproducibility|Average Time Required|Variance of Time Required|Log
---|---|---|---|---
Empty (disables GoReplay)|0%|-|-|[result/noseed.log](result/noseed.log)
Random ("1","2".."50")|36%|-|-|[result/randseed.log](result/randseed.log)
Fixed ("14")|14%|-|-|[result/fixedseed-14.log](result/fixedseed-14.log)
Fixed ("20")|50%|-|-|[result/fixedseed-20.log](result/fixedseed-20.log)

* To be documented: average/variance time

Fixed ("14") is not so reproducible due to poor injection points.
Fixed ("20") is much more reproducible, but its reproducibility seems achieved by just a long running time rather than GoReplay.

### Conclusion: the race is reproducible with GoReplay, but the complete execution is not so replayable yet
GoReplay can reproduce [etcd #5155](https://github.com/coreos/etcd/issues/5155) (data race) when the seed value varies.

But even with a fixed seed, the reproducibility doesn't so increase.
Pehaps we need to put more injection points and context information.
