package cola

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

type Client struct {
	BaseUrl  string
	Username string
	Password string
	Session  string
	Offset   int64
}

func (c *Client) Login() {
	form := map[string]interface{}{
		"username": c.Username,
		"password": c.Password,
	}
	url := c.BaseUrl + "/csras_api/login"
	param, err := json.Marshal(&form)
	if err != nil {
		log.Println("E!", "cola", err.Error())
		return
	}

	httpclient := &http.Client{Transport: tr}
	resp, err := httpclient.Get(url + "?param=" + string(param))
	if err != nil {
		log.Println("E!", "cola", err.Error())
		return
	}
	rv, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("E! ", "cola", err.Error())
		return
	}
	defer resp.Body.Close()

	msg := make(map[string]interface{})
	err = json.Unmarshal(rv, &msg)
	if err != nil {
		log.Println("E! ", "cola", err.Error())
		return
	}
	c.Session = msg["session"].(string)
}

func (c *Client) Logout() {
	if c.Session == "" {
		return
	}
	url := fmt.Sprintf(c.BaseUrl+"/csras_api/%s/logout", c.Session)
	httpclient := &http.Client{Transport: tr}
	resp, err := httpclient.Get(url)
	if err != nil {
		log.Println("E!", "cola", err.Error())
		return
	}
	defer resp.Body.Close()
}

func (c *Client) GetStatistic(netlink int, table string, fields []string, filter string) []map[string]interface{} {
	if c.Session == "" {
		return nil
	}

	begintime := time.Now().Unix()*1000 - c.Offset
	endtime := begintime + 1000
	data := map[string]interface{}{
		"netlink": netlink,
		"table":   table,
		"fields":  fields,
		"filter":  filter,

		"timeunit":   1000,
		"begintime":  begintime,
		"endtime":    endtime,
		"keys":       []string{"time"},
		"sorttype":   0,
		"sortfield":  "",
		"topcount":   1000,
		"keycount":   nil,
		"fieldcount": nil,
	}
	param, err := json.Marshal(&data)
	if err != nil {
		fmt.Println("E!", "cola", err.Error())
		return nil
	}

	url := fmt.Sprintf(c.BaseUrl+"/csras_api/%s/stats_data?param=%s", c.Session, string(param))
	httpclient := &http.Client{Transport: tr}

	resp, err := httpclient.Get(url)
	if err != nil {
		log.Println("E!", "cola", err.Error())
		return nil
	}
	defer resp.Body.Close()

	rv, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("E! ", "cola", err.Error())
		return nil
	}
	return parseData(&rv)
}

type Field struct {
	Name    string
	ValType uint8
}

func parseData(data *[]byte) []map[string]interface{} {
	reader := bytes.NewReader(*data)

	//解析返回结果
	var errcode int16
	binary.Read(reader, binary.BigEndian, &errcode)
	errmsg := readString(reader)
	if 0 != errcode {
		fmt.Println("errcode", errcode)
		fmt.Println("errmsg", errmsg)
	}

	//解析字段列表
	fields := readTableFields(reader)

	//跳过链路信息
	reader.Seek(1, io.SeekCurrent) //链路数量
	reader.Seek(2, io.SeekCurrent) //链路ID

	//记录个数
	var metrics = make([]map[string]interface{}, 0)
	var bucketCount int32
	binary.Read(reader, binary.BigEndian, &bucketCount)
	for i := 0; i < int(bucketCount); i++ {
		//不解析时间
		reader.Seek(8, io.SeekCurrent)

		//解析每一条记录
		var recordNum uint32
		binary.Read(reader, binary.BigEndian, &recordNum)
		for n := 0; n < int(recordNum); n++ {
			row := make(map[string]interface{})
			for _, field := range fields {
				if 1 == field.ValType {
					var v uint8
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = v
				} else if 2 == field.ValType {
					var v uint16
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = v
				} else if 3 == field.ValType {
					var v uint32
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = v
				} else if 4 == field.ValType {
					var v uint64
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = v
				} else if 5 == field.ValType {
					var v float64
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = v
				} else if 6 == field.ValType {
					var v int64
					binary.Read(reader, binary.BigEndian, &v)
					row[field.Name] = time.UnixMilli(v)
				} else if 7 == field.ValType {
					row[field.Name] = readString(reader)
				} else if 8 == field.ValType {
					var v uint32
					binary.Read(reader, binary.BigEndian, &v)
					percent := float32(v) / 100
					row[field.Name] = fmt.Sprintf("%f%%", percent)
				} else if 9 == field.ValType {
					var v = make([]byte, 8)
					reader.Read(v)
					var mac net.HardwareAddr = v
					row[field.Name] = mac.String()
				} else if 10 == field.ValType {
					var ip_ver uint8
					binary.Read(reader, binary.BigEndian, &ip_ver)

					if 4 == ip_ver {
						ipv4 := make([]byte, 4)
						reader.Read(ipv4)
						var ip net.IP = ipv4
						row[field.Name] = ip.String()
					}
					if 6 == ip_ver {
						ipv6 := make([]byte, 16)
						reader.Read(ipv6)
						var ip net.IP = ipv6
						row[field.Name] = ip.String()
					}
				} else {
					row[field.Name] = "N/A"
				}
			}
			metrics = append(metrics, row)
		}
	}

	// fmt.Println(metrics)
	return metrics
}

// ***************************************************************
// type           description                 bytes
//  1             UNIT8                         1
//  2             UNIT16                        2
//  3             UNIT32                        4
//  4             UINT64                        8
//  5             DOUBLE                        8
//  6             DATETIME                      8
//  7             TEXT                          8
//  8             PERCENT                       8
//  9             MAC                           8
//  10            IPADDR                      1+iplen
// ***************************************************************

func readString(reader *bytes.Reader) string {
	var size uint32
	binary.Read(reader, binary.BigEndian, &size)
	data := make([]byte, size)
	reader.Read(data)
	if data[len(data)-1] == 0 {
		data = data[0 : len(data)-1]
	}
	return string(data)
}

func readTableFields(reader *bytes.Reader) []Field {
	var len uint16
	binary.Read(reader, binary.BigEndian, &len)
	fields := make([]Field, len)
	for i := 0; i < int(len); i++ {
		fieldName := readString(reader)
		var valType uint8
		binary.Read(reader, binary.BigEndian, &valType)
		field := Field{Name: fieldName, ValType: valType}
		fields[i] = field
	}
	return fields
}
