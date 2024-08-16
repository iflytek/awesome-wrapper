package dfs

const MAXN = 10

var (
	graph          [MAXN][MAXN]int
	match          [MAXN]int
	visitX, visitY [MAXN]int
	nx, ny         int
)

func arrReset(arr []int, val int) {
	for ix := range arr {
		arr[ix] = val
	}
}
func findPath(u int) bool {
	visitX[u] = 1
	for v := 0; v < ny; v++ {
		if visitY[u] == 0 && graph[u][v] != 0 {
			visitY[v] = 1
			if match[v] == -1 || findPath(match[v]) {
				match[v] = u
				return true
			}
		}
	}
	return false
}

func dfsHungarian() int {
	var res = 0
	for i := 0; i < nx; i++ {
		arrReset(visitX[:], 0)
		arrReset(visitY[:], 0)
		if findPath(i) {
			res++
		}
	}
	return res
}
