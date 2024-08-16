package bfs

const MAXN = 10

var (
	graph                       [MAXN][MAXN]int
	matchX, matchY, prevX, chkY [MAXN]int
	queue                       [MAXN]int
	nx, ny                      int
)

func arrReset(arr []int, val int) {
	for ix := range arr {
		arr[ix] = val
	}
}

func bfsHungarian() int {
	var res = 0
	var qs, qe int
	arrReset(matchX[:], -1)
	arrReset(matchY[:], -1)
	arrReset(chkY[:], -1)

	for i := 0; i < nx; i++ {
		if matchX[i] == -1 {
			qs, qe = 0, 0
			queue[qe] = i
			qe += 1
			prevX[i] = -1
			flag := false
			for qs < qe && !flag {
				u := queue[qs]
				for v := 0; v < ny && !flag; v++ {
					if graph[u][v] != 0 && chkY[v] != i {
						chkY[v] = i
						queue[qe] = matchY[v]
						qe += 1
						if matchY[v] >= 0 {
							prevX[matchY[v]] = u
						} else {
							flag = true
							var d, e = u, v
							for d != -1 {
								t := matchX[d]
								matchX[d] = e
								matchY[e] = d
								d = prevX[d]
								e = t
							}
						}
					}
				}
				qs++
			}
			if matchX[i] != -1 {
				res += 1
			}
		}
	}
	return res
}
