package httpserver

import (
	"net/http"
	"strings"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func (s *HttpServer) UpdateHandler(w http.ResponseWriter, req *http.Request) {

	// lets' check method first
	if req.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// let's split URl into parts to check if request is corect
	urlParts := strings.Split(req.URL.Path, "/")

	// url should look like /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	// so if there are less than 5 parts, request is not correct
	if len(urlParts) < 5 {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// part 0 should be empty and part 1 should be "update", just in case
	if urlParts[0] != "" || urlParts[1] != "update" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	metricType := urlParts[2]
	metricName := urlParts[3]
	metricValue := urlParts[4]

	m, err := s.Storage.Retrieve(metrics.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		m, err = metrics.NewMetric(metrics.MetricType(metricType), metricName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = s.Storage.Add(m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = m.Update(metricValue)
	if err != nil {

		if err.Error() == metrics.ErrorInvalidMetricValue {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// all, err := s.Storage.RetrieveAll()

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// fmt.Println("")
	// fmt.Println("")
	// fmt.Println("====")
	// for a, b := range all {
	// 	fmt.Println(a, b)
	// }

	// if everything is correct and metric was saved
	w.Write([]byte("OK"))
}
