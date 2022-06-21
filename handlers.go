package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func toInt(val string) int {
	value, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println(err)
		return 0
	} else {
		return value
	}
}

func toString(val int) string {
	return strconv.Itoa(val)
}

func runSql(s string) *sql.Rows {
	var result *sql.Rows
	r, err := db.Query(s)
	if err != nil {
		fmt.Println(err)
	}
	result = r
	return result
}

func runSqlChannel(s string) <-chan *sql.Rows {
	ch := make(chan *sql.Rows)
	go func() {
		defer close(ch)
		r, err := db.Query(s)
		if err != nil {
			fmt.Println(err)
		}
		ch <- r
	}()
	return ch
}

func getDataWorldwide(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var url = "/data/worldwide"
	if limit != "" {
		limitInt = toInt(limit)
		//limitInt++
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/worldwide" {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	//vars
	var dataWorldwide = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dwws []DataCountriesDeltaPercent
	var result string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	result, err = rdb.Get(url).Result()

	//main
	if err == redis.Nil {
		//sql
		result, err := db.Query("SELECT * FROM data.v_all_delta_worldwide ORDER BY date DESC LIMIT ? OFFSET ?", limitPlus, offset)
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		//result
		for result.Next() {
			var dww DataCountriesDeltaPercent
			err := result.Scan(&dww.Country, &dww.Date, &dww.Confirmed, &dww.DeltaConfirmed, &dww.DeltaConfirmedPercent, &dww.Deaths, &dww.DeathsPercent, &dww.DeltaDeaths, &dww.DeltaDeathsPercent, &dww.Recovered, &dww.RecoveredPercent, &dww.DeltaRecovered, &dww.DeltaRecoveredPercent, &dww.Active, &dww.ActivePercent, &dww.DeltaActive, &dww.DeltaActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dwws = append(dwws, dww)
		}
		if len(dwws) > limitInt {
			dataWorldwide["data"] = dwws[:len(dwws)-1]
			meta["hasMore"] = true
		} else {
			dataWorldwide["data"] = dwws
			meta["hasMore"] = false
		}
		dataWorldwide["meta"] = meta
		dataWorldwideJson, err := json.Marshal(dataWorldwide)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataWorldwideJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataWorldwide)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(result), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataCountriesLast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var sortKey = r.URL.Query().Get("sortKey")
	var order = r.URL.Query().Get("order")
	var url = "/data/countries"
	if limit != "" {
		limitInt = toInt(limit)
		//limitInt++
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/countries" {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	if sortKey != "" {
		if url == "/data/countries" {
			url = url + "?sortKey=" + sortKey
		} else {
			url = url + "&sortKey=" + sortKey
		}
	}
	if order != "" {
		if url == "/data/countries" {
			url = url + "?order=" + order
		} else {
			url = url + "&order=" + order
		}
	}
	//vars
	var dataCountriesLast = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dcss []DataCountries
	var res string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	res, err = rdb.Get(url).Result()

	//main

	if err == redis.Nil {
		//sql
		var sqlData = fmt.Sprintf("SELECT * FROM data.v_all_by_country ORDER BY %s %s LIMIT %s OFFSET %s", sortKey, order, limitPlus, offset)
		var sqlMeta = fmt.Sprintf("SELECT COUNT(*) as total FROM data.v_all_by_country")

		var result, pages = <-runSqlChannel(sqlData), <-runSqlChannel(sqlMeta)

		for result.Next() {
			var dcs DataCountries
			err := result.Scan(&dcs.Country, &dcs.Date, &dcs.Confirmed, &dcs.Deaths, &dcs.DeathsPercent, &dcs.Recovered, &dcs.RecoveredPercent, &dcs.Active, &dcs.ActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dcss = append(dcss, dcs)
		}

		var total int
		for pages.Next() {
			err := pages.Scan(&total)
			if err != nil {
				panic(err.Error())
			}
		}

		if len(dcss) > limitInt {
			dataCountriesLast["data"] = dcss[:len(dcss)-1]
			meta["hasMore"] = true
		} else {
			dataCountriesLast["data"] = dcss
			meta["hasMore"] = false
		}
		meta["pages"] = math.Floor(float64(total)/float64(limitInt)) + 1
		dataCountriesLast["meta"] = meta
		dataCountriesLastJson, err := json.Marshal(dataCountriesLast)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataCountriesLastJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataCountriesLast)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(res), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataCountriesDeltaLast(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var sortKey = r.URL.Query().Get("sortKey")
	var order = r.URL.Query().Get("order")
	var url = "/data/countries/delta/last"
	if limit != "" {
		limitInt = toInt(limit)
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/countries/delta/last" {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	if sortKey != "" {
		if url == "/data/countries/delta/last" {
			url = url + "?sortKey=" + sortKey
		} else {
			url = url + "&sortKey=" + sortKey
		}
	}
	if order != "" {
		if url == "/data/countries/delta/last" {
			url = url + "?order=" + order
		} else {
			url = url + "&order=" + order
		}
	}
	//vars
	var dataCountriesDeltaLast = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dcss []DataCountriesDelta
	var res string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	res, err = rdb.Get(url).Result()

	//main

	if err == redis.Nil {
		//sql
		var sqlData = fmt.Sprintf("SELECT * FROM data.v_delta_1d_by_country ORDER BY %s %s LIMIT %s OFFSET %s", sortKey, order, limitPlus, offset)
		var sqlMeta = fmt.Sprintf("SELECT COUNT(*) as total FROM data.v_delta_1d_by_country")

		var result, pages = <-runSqlChannel(sqlData), <-runSqlChannel(sqlMeta)

		for result.Next() {
			var dcs DataCountriesDelta
			err := result.Scan(&dcs.Country, &dcs.Date, &dcs.Confirmed, &dcs.DeltaConfirmed, &dcs.DeltaConfirmedPercent, &dcs.Deaths, &dcs.DeltaDeaths, &dcs.DeltaDeathsPercent, &dcs.Recovered, &dcs.DeltaRecovered, &dcs.DeltaRecoveredPercent, &dcs.Active, &dcs.DeltaActive, &dcs.DeltaActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dcss = append(dcss, dcs)
		}

		var total int
		for pages.Next() {
			err := pages.Scan(&total)
			if err != nil {
				panic(err.Error())
			}
		}

		if len(dcss) > limitInt {
			dataCountriesDeltaLast["data"] = dcss[:len(dcss)-1]
			meta["hasMore"] = true
		} else {
			dataCountriesDeltaLast["data"] = dcss
			meta["hasMore"] = false
		}
		meta["pages"] = math.Floor(float64(total)/float64(limitInt)) + 1
		dataCountriesDeltaLast["meta"] = meta
		dataCountriesDeltaLastJson, err := json.Marshal(dataCountriesDeltaLast)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataCountriesDeltaLastJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataCountriesDeltaLast)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(res), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataConfirmed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var url = "/data/countries/confirmed"

	//vars
	var dataConfirmed = make(map[string]interface{})
	var dc []DataConfirmed
	var result string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	result, err = rdb.Get(url).Result()

	//main
	if err == redis.Nil {
		//sql
		result, err := db.Query("SELECT * FROM data.v_confirmed_by_country_percent")
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		//result
		for result.Next() {
			var d DataConfirmed
			err := result.Scan(&d.Country, &d.Date, &d.Confirmed, &d.ConfirmedPercent, &d.TotalConfirmed)
			if err != nil {
				panic(err.Error())
			}
			dc = append(dc, d)
		}
		dataConfirmed["data"] = dc
		dataConfirmedJson, err := json.Marshal(dataConfirmed)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataConfirmedJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataConfirmed)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(result), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataDeltaConfirmed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var url = "/data/countries/confirmed/delta"

	//vars
	var dataDeltaConfirmed = make(map[string]interface{})
	var dc []DataDeltaConfirmed
	var result string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	result, err = rdb.Get(url).Result()

	//main
	if err == redis.Nil {
		//sql
		result, err := db.Query("SELECT * FROM data.v_delta_1d_confirmed_by_country_percent;")
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		//result
		for result.Next() {
			var d DataDeltaConfirmed
			err := result.Scan(&d.Country, &d.Date, &d.Confirmed, &d.DeltaConfirmed, &d.DeltaConfirmedPercent, &d.DeltaTotalConfirmed)
			if err != nil {
				panic(err.Error())
			}
			dc = append(dc, d)
		}
		dataDeltaConfirmed["data"] = dc
		dataDeltaConfirmedJson, err := json.Marshal(dataDeltaConfirmed)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataDeltaConfirmedJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataDeltaConfirmed)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(result), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataDeltaActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var url = "/data/countries/active/delta"
	if limit != "" {
		limitInt = toInt(limit)
		//limitInt++
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/countries/active/delta" {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	//vars
	var dataDeltaActive = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dda []DataDeltaActive
	var result string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	result, err = rdb.Get(url).Result()

	//main
	if err == redis.Nil {
		//sql
		result, err := db.Query("SELECT * FROM data.v_delta_1d_active_by_country WHERE delta_active < 0 ORDER BY delta_active ASC LIMIT ? OFFSET ?", limitPlus, offset)
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		//result
		for result.Next() {
			var d DataDeltaActive
			err := result.Scan(&d.Country, &d.Date, &d.Confirmed, &d.Deaths, &d.Recovered, &d.Active, &d.DeltaActive, &d.DeltaActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dda = append(dda, d)
		}
		if len(dda) > limitInt {
			dataDeltaActive["data"] = dda[:len(dda)-1]
			meta["hasMore"] = true
		} else {
			dataDeltaActive["data"] = dda
			meta["hasMore"] = false
		}
		dataDeltaActive["meta"] = meta
		dataDeltaActiveJson, err := json.Marshal(dataDeltaActive)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataDeltaActiveJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataDeltaActive)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(result), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataCountriesList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//query
	var url = "/data/countries/list"

	//vars
	var dataCountriesList = make(map[string]interface{})
	var dc []DataCountriesList
	var result string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	result, err = rdb.Get(url).Result()

	//main
	if err == redis.Nil {
		//sql
		result, err := db.Query("SELECT distinct(country) FROM data.data_confirmed ORDER BY country ASC")
		if err != nil {
			panic(err.Error())
		}
		defer result.Close()
		//result
		for result.Next() {
			var d DataCountriesList
			err := result.Scan(&d.Country)
			if err != nil {
				panic(err.Error())
			}
			dc = append(dc, d)
		}
		dataCountriesList["data"] = dc
		dataCountriesListJson, err := json.Marshal(dataCountriesList)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataCountriesListJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataCountriesList)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(result), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataCountries(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var sortKey = r.URL.Query().Get("sortKey")
	var order = r.URL.Query().Get("order")
	var url = "/data/countries/" + params["country"]

	if limit != "" {
		limitInt = toInt(limit)
		//limitInt++
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/countries"+params["country"] {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	if sortKey != "" {
		if url == "/data/countries"+params["country"] {
			url = url + "?sortKey=" + sortKey
		} else {
			url = url + "&sortKey=" + sortKey
		}
	}
	if order != "" {
		if url == "/data/countries"+params["country"] {
			url = url + "?order=" + order
		} else {
			url = url + "&order=" + order
		}
	}
	//vars
	var dataCountries = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dcss []DataCountries
	var res string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	res, err = rdb.Get(url).Result()

	//main

	if err == redis.Nil {
		//sql
		var sqlData = fmt.Sprintf("SELECT * FROM data.v_data_by_country WHERE country = '%s' ORDER BY %s %s LIMIT %s OFFSET %s", params["country"], sortKey, order, limitPlus, offset)
		var sqlMeta = fmt.Sprintf("SELECT COUNT(*) as total FROM data.v_data_by_country WHERE country ='%s'", params["country"])

		var result, pages = <-runSqlChannel(sqlData), <-runSqlChannel(sqlMeta)

		for result.Next() {
			var dcs DataCountries
			err := result.Scan(&dcs.Country, &dcs.Date, &dcs.Confirmed, &dcs.Deaths, &dcs.DeathsPercent, &dcs.Recovered, &dcs.RecoveredPercent, &dcs.Active, &dcs.ActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dcss = append(dcss, dcs)
		}

		var total int
		for pages.Next() {
			err := pages.Scan(&total)
			if err != nil {
				panic(err.Error())
			}
		}

		if len(dcss) > limitInt {
			dataCountries["data"] = dcss[:len(dcss)-1]
			meta["hasMore"] = true
		} else {
			dataCountries["data"] = dcss
			meta["hasMore"] = false
		}
		meta["pages"] = math.Floor(float64(total)/float64(limitInt)) + 1
		dataCountries["meta"] = meta
		dataCountriesJson, err := json.Marshal(dataCountries)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataCountriesJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataCountries)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(res), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}

func getDataCountriesDelta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var params = mux.Vars(r)
	//query
	var limit = r.URL.Query().Get("limit")
	var limitPlus string
	var limitInt int
	var offset = r.URL.Query().Get("offset")
	var sortKey = r.URL.Query().Get("sortKey")
	var order = r.URL.Query().Get("order")
	var url = "/data/countries/" + params["country"] + "/delta"

	if limit != "" {
		limitInt = toInt(limit)
		limitPlus = toString(limitInt + 1)
		url = url + "?limit=" + limit
	}
	if offset != "" {
		if url == "/data/countries"+params["country"]+"/delta" {
			url = url + "?offset=" + offset
		} else {
			url = url + "&offset=" + offset
		}
	}
	if sortKey != "" {
		if url == "/data/countries"+params["country"]+"/delta" {
			url = url + "?sortKey=" + sortKey
		} else {
			url = url + "&sortKey=" + sortKey
		}
	}
	if order != "" {
		if url == "/data/countries"+params["country"]+"/delta" {
			url = url + "?order=" + order
		} else {
			url = url + "&order=" + order
		}
	}
	//vars
	var dataCountriesDelta = make(map[string]interface{})
	var meta = make(map[string]interface{})
	var dcss []DataCountriesDelta
	var res string
	//redis
	rdb := newRedisClient()
	defer rdb.Close()
	res, err = rdb.Get(url).Result()

	//main

	if err == redis.Nil {
		//sql
		var sqlData = fmt.Sprintf("SELECT * FROM data.v_delta_data_by_country WHERE country = '%s' ORDER BY %s %s LIMIT %s OFFSET %s", params["country"], sortKey, order, limitPlus, offset)
		var sqlMeta = fmt.Sprintf("SELECT COUNT(*) as total FROM data.v_delta_data_by_country WHERE country ='%s'", params["country"])

		var result, pages = <-runSqlChannel(sqlData), <-runSqlChannel(sqlMeta)

		for result.Next() {
			var dcs DataCountriesDelta
			err := result.Scan(&dcs.Country, &dcs.Date, &dcs.Confirmed, &dcs.DeltaConfirmed, &dcs.DeltaConfirmedPercent, &dcs.Deaths, &dcs.DeltaDeaths, &dcs.DeltaDeathsPercent, &dcs.Recovered, &dcs.DeltaRecovered, &dcs.DeltaRecoveredPercent, &dcs.Active, &dcs.DeltaActive, &dcs.DeltaActivePercent)
			if err != nil {
				panic(err.Error())
			}
			dcss = append(dcss, dcs)
		}

		var total int
		for pages.Next() {
			err := pages.Scan(&total)
			if err != nil {
				panic(err.Error())
			}
		}

		if len(dcss) > limitInt {
			dataCountriesDelta["data"] = dcss[:len(dcss)-1]
			meta["hasMore"] = true
		} else {
			dataCountriesDelta["data"] = dcss
			meta["hasMore"] = false
		}
		meta["pages"] = math.Floor(float64(total)/float64(limitInt)) + 1
		dataCountriesDelta["meta"] = meta
		dataCountriesDeltaJson, err := json.Marshal(dataCountriesDelta)
		if err != nil {
			log.Fatal("Cannot encode to JSON ", err)
		}
		err = rdb.Set(url, dataCountriesDeltaJson, 3600*time.Second).Err()
		if err != nil {
			panic(err)
		}
		json.NewEncoder(w).Encode(dataCountriesDelta)

	} else if err != nil {
		fmt.Println("Unknown error")
	} else {
		var i interface{}
		json.Unmarshal([]byte(res), &i)
		if err != nil {
			log.Fatalf("failed to decode: %s", err)
		}
		json.NewEncoder(w).Encode(i)
	}
}
