package daemon

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

//RDBMS(Relational Database Management System
type RowData struct {
	segIdDb    string
	typeDb     string
	serverIpDb string
}

var columnJson = map[string]string{
	"seg_id":    "seg_id",
	"type":      "type",
	"server_ip": "server_ip",
}

type oriStruct struct {
	Ret    string `json:"ret"`
	Result []struct {
		SegID    string `json:"seg_id"`
		ServerIP string `json:"server_ip"`
		Type     string `json:"type"`
	} `json:"result"`
}

var columnJsonRwMu sync.RWMutex

var MysqlManagerInst MysqlManager

type MysqlManager struct {
	WebService
}

func (m *MysqlManager) Init(baseUrl string, caller, callerKey string, timeout time.Duration, token string, version string, idc string, schema string, table string) (bool, error) {
	return m.WebService.Init(baseUrl, caller, callerKey, timeout, token, version, idc, schema, table)
}

/*
1、将原始的行结构形式的数据处理为map形式
2、k为seg_id，v为server_ip
*/
func (m *MysqlManager) GetSubSvcSegIdData(subSvc string) (rst map[string]string, err error) {
	if rst == nil {
		rst = make(map[string]string)
	}
	rows, rowsErr := m.Retrieve(map[string]string{"type": subSvc})
	if rowsErr != nil {
		err = rowsErr
		return
	}
	for _, v := range rows {
		rst[v.segIdDb] = v.serverIpDb
	}
	return
}
func (m *MysqlManager) AddNewSegIdData(row RowData) (bool, error) {
	return m.Insert(map[string]string{"seg_id": row.segIdDb, "type": row.typeDb, "server_ip": row.serverIpDb})
}
func (m *MysqlManager) DelServer(addr string) (bool, error) {
	return m.Delete(map[string]string{"server_ip": addr})
}
func (m *MysqlManager) db2Rows(in string) (rst []RowData, err error) {
	tmp := oriStruct{}
	if err := json.Unmarshal([]byte(in), &tmp); err != nil {
		return nil, fmt.Errorf("can't parse the json -> %v,err -> %v", in, err)
	}
	if tmp.Ret != "0" {
		return nil, fmt.Errorf("the ret -> %v not equal 0", tmp.Ret)
	}
	for _, v := range tmp.Result {
		rst = append(rst, RowData{segIdDb: v.SegID, typeDb: v.Type, serverIpDb: v.ServerIP})
	}
	return rst, nil
}

func (m *MysqlManager) Retrieve(where map[string]string) ([]RowData, error) {
	columnJsonRwMu.RLock()
	defer columnJsonRwMu.RUnlock()
	GetListStr, GetListStrErr := m.WebService.GetList(columnJson, where)
	if GetListStrErr != nil {
		return nil, GetListStrErr
	}
	return m.db2Rows(GetListStr)
}

func (m *MysqlManager) Create(increase []RowData) bool {
	panic("implement me")
}

func (m *MysqlManager) Update(set RowData, where map[string]string) bool {
	panic("implement me")
}

func (m *MysqlManager) Delete(where map[string]string) (bool, error) {
	return m.WebService.Delete(where)
}
