// Copyright 2014 The dbrouter Author. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


package dbrouter

import (
	"fmt"
	//"sync"
	"encoding/json"
)

const (
	DB_TYPE_MONGO = "mongo"
	DB_TYPE_MYSQL = "mysql"
)

type dbLookupCfg struct {
	Instance string  `json:"instance"`
	// match type: full or regex
	Match string     `json:"match"`
	Express string   `json:"express"`

}

func (m *dbLookupCfg) String() string {
	return fmt.Sprintf("ins:%s exp:%s match:%s", m.Instance, m.Express, m.Match)
}


type dbInsCfg struct {
	Dbtype string             `json:"dbtype"`
	Dbname string             `json:"dbname"`
	//Dbcfg map[string]interface{}   `json:"dbcfg"`
	Dbcfg json.RawMessage      `json:"dbcfg"`
}


type routeConfig struct {
	Cluster map[string][]*dbLookupCfg `json:"cluster"`
	Instances map[string] *dbInsCfg  `json:"instances"`
}

type Router struct {
	dbCls *dbCluster
	dbIns *dbInstanceManager
}

func (m *Router) String() string {
	return fmt.Sprintf("%s", m.dbCls.clusters)
}


func NewRouter(jscfg []byte) (*Router, error) {
	var cfg routeConfig
	err := json.Unmarshal(jscfg, &cfg)
	if err != nil {
		return nil, fmt.Errorf("dbrouter config unmarshal:%s", err)
	}



	r := &Router {
		dbCls: &dbCluster {
			clusters: make(map[string]*clsEntry),
		},

		dbIns: &dbInstanceManager {
			instances: make(map[string]dbInstance),

		},
	}

	cls := cfg.Cluster
	for c, ins := range cls {
		for _, v := range ins {
			if err := r.dbCls.addInstance(c, v); err != nil {
				return nil, fmt.Errorf("load instance lookup rule err:%s", err.Error())
			}
		}
	}

	inss := cfg.Instances
	for ins, db := range inss {
		tp := db.Dbtype
		dbname := db.Dbname
		cfg := db.Dbcfg
		// 工厂化构造，db类型领出来
		if tp == DB_TYPE_MONGO {
			dbi, err := NewdbMongo(tp, dbname, cfg)
			if err != nil {
				return nil, fmt.Errorf("init mongo config err:%s", err.Error())
			}

			r.dbIns.add(ins, dbi)
		}

	}

	return r, nil
}



// 通过传入配置方式加载，配置的结构对外面隐藏
// 无论是全匹配，还是正则匹配，被查找的表明必须全部被匹配命中才能生效
// 全匹配优先进行
// db cfg虽然是透传，但是也增加json检查??
// 更细节的err输出,不要只单独返回err，还要返回时哪里的err
