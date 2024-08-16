/*==================================================*\
 |  二分图匹配（匈牙利算法DFS 实现）
 |  INIT: graph[][]邻接矩阵;
 |  CALL: res =  dfsHungarian ();
 |  优点：实现简洁容易理解，适用于稠密图，DFS 找增广路快。
 |  找一条增广路的复杂度为O(E)，最多找V条增广路，故时间复杂度为O(VE)
 |  算法简述：
 |  从二分图中找出一条路径来，让路径的起点和终点都是还没有匹配过的点，
 |  并且路径经过的连线是一条没被匹配、一条已经匹配过，再下一条又没匹配这样交替地出现。
 |  找到这样的路径后，显然路径里没被匹配的连线比已经匹配了的连线多一条，
 |  于是修改匹配图，把路径里所有匹配过的连线去掉匹配关系，把没有匹配的连线变成匹配的。
 |  这样匹配数就比原来多1个。不断执行上述操作，直到找不到这样的路径为止。
 \*==================================================*/
#include<iostream>
#include<memory.h>
using namespace std;

#define MAXN 10
int graph[MAXN][MAXN];
int match[MAXN];
int visitX[MAXN], visitY[MAXN];
int nx, ny;

bool findPath( int u )
{
    visitX[u] = 1;
    for( int v=0; v<ny; v++ )
    {
        if( !visitY[v] && graph[u][v] )
        {
            visitY[v] = 1;
            if( match[v] == -1 //第一次,用到了短路计算，否则findPath(-1)会出问题
                    || findPath(match[v]) )//这里就表示深度优先遍历 不撞南山头不回 不见黄河心不死
            {
                match[v] = u;
                return true;
            }
        }
    }
    return false;
}

int dfsHungarian()
{
    int res = 0;
    memset( match, -1, sizeof(match) );
    for( int i=0; i<nx; i++ )
    {
        memset( visitX, 0, sizeof(visitX) );
        memset( visitY, 0, sizeof(visitY) );
        if( findPath(i) )
            res++;
    }
    return res;
}