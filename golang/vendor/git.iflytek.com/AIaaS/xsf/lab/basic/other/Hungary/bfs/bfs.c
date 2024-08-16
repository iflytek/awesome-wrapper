/*==================================================*\
  |  二分图匹配（匈牙利算法BFS 实现）
  |  INIT: graph[][]邻接矩阵;
  |  CALL: res =  bfsHungarian ();
  |  优点：适用于稀疏二分图，边较少，增广路较短。
  |  匈牙利算法的理论复杂度是O(VE)
  \*==================================================*/
#include<iostream>
#include<memory.h>
using namespace std;

#define MAXN 10
int graph[MAXN][MAXN];
//在bfs中，增广路径的搜索是一层一层展开的，所以必须通过prevX来记录上一层的顶点
//chkY用于标记某个Y顶点是否被目前的X顶点访问尝试过。
int matchX[MAXN], matchY[MAXN], prevX[MAXN], chkY[MAXN];
int queue[MAXN];
int nx, ny;

int bfsHungarian()
{
    int res = 0;
    int qs, qe;
    memset( matchX, -1, sizeof(matchX) );
    memset( matchY, -1, sizeof(matchY) );
    memset( chkY, -1, sizeof(chkY) );

    for( int i=0; i<nx; i++ )
    {
        if( matchX[i] == -1 )   //如果该X顶点未找到匹配点，将其放入队列。
        {
            qs = qe = 0;
            queue[qe++] = i;
            prevX[i] = -1;  //并且标记，它是路径的起点
            bool flag = 0;

            while( qs<qe && !flag )
            {
                int u = queue[qs];
                for( int v=0; v<ny&&!flag; v++ )
                {
                    if( graph[u][v] && chkY[v]!=i ) //如果该节点与u有边且未被访问过
                    {
                        chkY[v] = i;    //标记且将它的前一个顶点放入队列中，也就是下次可能尝试这个顶点看能否为它找到新的节点
                        queue[qe++] = matchY[v];
                        if( matchY[v] >= 0 )
                            prevX[matchY[v]] = u;
                        else    //到达了增广路径的最后一站
                        {
                            flag = 1;
                            int d=u, e=v;
                            while( d!=-1 )  //一路通过prevX找到路径的起点
                            {
                                int t = matchX[d];
                                matchX[d] = e;
                                matchY[e] = d;
                                d = prevX[d];
                                e = t;
                            }
                        }
                    }
                }
                qs++;
            }
            if( matchX[i] != -1 )
                res++;
        }
    }
    return res;
}