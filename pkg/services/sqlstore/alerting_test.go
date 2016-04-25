package sqlstore

import (
	"testing"

	m "github.com/grafana/grafana/pkg/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestAlertingDataAccess(t *testing.T) {

	Convey("Testing Alerting data access", t, func() {
		InitTestDB(t)

		testDash := insertTestDashboard("dashboard with alerts", 1, "alert")

		items := []m.AlertRule{
			{
				PanelId:     1,
				DashboardId: testDash.Id,
				OrgId:       testDash.OrgId,
				Query:       "Query",
				QueryRefId:  "A",
				WarnLevel:   "> 30",
				CritLevel:   "> 50",
				Interval:    "10",
				Title:       "Alerting title",
				Description: "Alerting description",
				QueryRange:  "5m",
				Aggregator:  "avg",
			},
		}

		cmd := m.SaveAlertsCommand{
			Alerts:      &items,
			DashboardId: testDash.Id,
			OrgId:       1,
			UserId:      1,
		}

		err := SaveAlerts(&cmd)

		Convey("Can create one alert", func() {
			So(err, ShouldBeNil)

			alertChanges, er := GetAlertRuleChanges(1)
			So(er, ShouldBeNil)
			So(len(alertChanges), ShouldEqual, 1)
		})

		Convey("Can read properties", func() {
			alert, err2 := GetAlertsByDashboardAndPanelId(testDash.Id, 1)

			So(err2, ShouldBeNil)
			So(alert.Interval, ShouldEqual, "10")
			So(alert.WarnLevel, ShouldEqual, "> 30")
			So(alert.CritLevel, ShouldEqual, "> 50")
			So(alert.Query, ShouldEqual, "Query")
			So(alert.QueryRefId, ShouldEqual, "A")
			So(alert.Title, ShouldEqual, "Alerting title")
			So(alert.Description, ShouldEqual, "Alerting description")
			So(alert.QueryRange, ShouldEqual, "5m")
			So(alert.Aggregator, ShouldEqual, "avg")
		})

		Convey("Alerts with same dashboard id and panel id should update", func() {
			modifiedItems := items
			modifiedItems[0].Query = "Updated Query"

			modifiedCmd := m.SaveAlertsCommand{
				DashboardId: testDash.Id,
				OrgId:       1,
				UserId:      1,
				Alerts:      &modifiedItems,
			}

			err := SaveAlerts(&modifiedCmd)

			Convey("Can save alerts with same dashboard and panel id", func() {
				So(err, ShouldBeNil)
			})

			Convey("Alerts should be updated", func() {
				alerts, err2 := GetAlertsByDashboardId(testDash.Id)

				So(err2, ShouldBeNil)
				So(len(alerts), ShouldEqual, 1)
				So(alerts[0].Query, ShouldEqual, "Updated Query")
			})

			Convey("Updates without changes should be ignored", func() {
				err3 := SaveAlerts(&modifiedCmd)
				So(err3, ShouldBeNil)

				alertChanges, er := GetAlertRuleChanges(1)
				So(er, ShouldBeNil)
				So(len(alertChanges), ShouldEqual, 2)
			})
		})

		Convey("Multiple alerts per dashboard", func() {
			multipleItems := []m.AlertRule{
				{
					DashboardId: testDash.Id,
					PanelId:     1,
					Query:       "1",
					OrgId:       1,
				},
				{
					DashboardId: testDash.Id,
					PanelId:     2,
					Query:       "2",
					OrgId:       1,
				},
				{
					DashboardId: testDash.Id,
					PanelId:     3,
					Query:       "3",
					OrgId:       1,
				},
			}

			cmd.Alerts = &multipleItems
			err = SaveAlerts(&cmd)

			Convey("Should save 3 dashboards", func() {
				So(err, ShouldBeNil)

				alerts, err2 := GetAlertsByDashboardId(testDash.Id)
				So(err2, ShouldBeNil)
				So(len(alerts), ShouldEqual, 3)
				alertChanges, er := GetAlertRuleChanges(1)
				So(er, ShouldBeNil)
				So(len(alertChanges), ShouldEqual, 4)
			})

			Convey("should updated two dashboards and delete one", func() {
				missingOneAlert := multipleItems[:2]

				cmd.Alerts = &missingOneAlert
				err = SaveAlerts(&cmd)

				Convey("should delete the missing alert", func() {
					alerts, err2 := GetAlertsByDashboardId(testDash.Id)
					So(err2, ShouldBeNil)
					So(len(alerts), ShouldEqual, 2)
				})

				Convey("should add one more alert_rule_change", func() {
					alertChanges, er := GetAlertRuleChanges(1)
					So(er, ShouldBeNil)
					So(len(alertChanges), ShouldEqual, 6)
				})
			})
		})

		Convey("When dashboard is removed", func() {
			items := []m.AlertRule{
				{
					PanelId:     1,
					DashboardId: testDash.Id,
					Query:       "Query",
					QueryRefId:  "A",
					WarnLevel:   "> 30",
					CritLevel:   "> 50",
					Interval:    "10",
					Title:       "Alerting title",
					Description: "Alerting description",
					QueryRange:  "5m",
					Aggregator:  "avg",
				},
			}

			cmd := m.SaveAlertsCommand{
				Alerts:      &items,
				DashboardId: testDash.Id,
				OrgId:       1,
				UserId:      1,
			}

			SaveAlerts(&cmd)

			err = DeleteDashboard(&m.DeleteDashboardCommand{
				OrgId: 1,
				Slug:  testDash.Slug,
			})

			So(err, ShouldBeNil)

			Convey("Alerts should be removed", func() {
				alerts, err2 := GetAlertsByDashboardId(testDash.Id)

				So(testDash.Id, ShouldEqual, 1)
				So(err2, ShouldBeNil)
				So(len(alerts), ShouldEqual, 0)
			})
		})
	})
}