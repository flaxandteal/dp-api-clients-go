package areas

import (
	"context"
	"fmt"
	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	userAuthToken    = "iamatoken"
	serviceAuthToken = "iamaservicetoken"
	collectionID     = "iamacollectionID"
	testHost         = "http://localhost:8080"
)

var (
	ctx          = context.Background()
	initialState = health.CreateCheckState(service)
)

type MockedHTTPResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)

		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 500 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 500,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 404 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 404,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 404)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given clienter.Do returns 429 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 429,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, 429)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given clienter.Do returns 200 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{
			StatusCode: 200,
		}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		areasClient := newAreasClient(httpClient)
		check := initialState

		Convey("when areasClient.Checker is called", func() {
			err := areasClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure(), ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_GetArea(t *testing.T) {

	areaBody := `{
	"code": "E92000001",
	"name": "England",
	"date_started": "Thu, 01 Jan 2009 00:00:00 GMT",
	"date_end": "null",
	"name_welsh": "Lloegr",
	"limit": "1",
	"total_count": "10",
	"visible": true
	}`

	Convey("When bad request is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 400, Body: ""})
		_, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001")
		So(err, ShouldNotBeNil)
	})

	Convey("When a area is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: areaBody})
		area, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001")
		So(err, ShouldBeNil)
		So(area, ShouldResemble, AreaDetails{
			Code:        "E92000001",
			Name:        "England",
			DateStarted: "Thu, 01 Jan 2009 00:00:00 GMT",
			DateEnd:     "null",
			NameWelsh:   "Lloegr",
			Limit:       "1",
			TotalCount:  "10",
			Visible:     true,
		})
	})

	Convey("given a 200 status with valid empty body is returned", t, func() {
		mockedAPI := getMockAreaAPI(http.Request{Method: "GET"}, MockedHTTPResponse{StatusCode: 200, Body: "{}"})

		Convey("when GetArea is called", func() {
			instance, err := mockedAPI.GetArea(ctx, userAuthToken, serviceAuthToken, collectionID, "E92000001")

			Convey("a positive response is returned with empty instance", func() {
				So(err, ShouldBeNil)
				So(instance, ShouldResemble, AreaDetails{})
			})
		})
	})
}

func getMockAreaAPI(expectRequest http.Request, mockedHTTPResponse MockedHTTPResponse) *Client {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectRequest.Method {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected HTTP method used"))
			return
		}
		w.WriteHeader(mockedHTTPResponse.StatusCode)
		fmt.Fprintln(w, mockedHTTPResponse.Body)
	}))
	return New(ts.URL)
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newAreasClient(clienter *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, clienter)
	areasClient := NewWithHealthClient(healthClient)
	return areasClient
}