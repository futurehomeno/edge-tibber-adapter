package handler

import (
	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/fimptype"
	log "github.com/sirupsen/logrus"
	tibber "github.com/tskaard/tibber-golang"
)

func createInterface(iType string, msgType string, valueType string, version string) fimptype.Interface {
	inter := fimptype.Interface{
		Type:      iType,
		MsgType:   msgType,
		ValueType: valueType,
		Version:   version,
	}
	return inter
}

func createSensorService(addr string, service string, supUnits []string, alias string) fimptype.Service {
	cmdSensorGetReport := createInterface("in", "cmd.sensor.get_report", "null", "1")
	evtSensorReport := createInterface("out", "evt.sensor.report", "float", "1")
	sensorInterfaces := []fimptype.Interface{}
	sensorInterfaces = append(sensorInterfaces, cmdSensorGetReport, evtSensorReport)

	props := make(map[string]interface{})
	props["sup_units"] = supUnits
	sensorService := fimptype.Service{
		Address:    "/rt:dev/rn:tibber/ad:1/sv:" + service + "/ad:" + addr,
		Name:       service,
		Groups:     []string{"ch_0"},
		Alias:      alias,
		Enabled:    true,
		Props:      props,
		Interfaces: sensorInterfaces,
	}
	return sensorService
}

func (t *FimpTibberHandler) sendInclusionReport(home tibber.Home, oldMsg *fimpgo.FimpMessage) {
	services := []fimptype.Service{}

	powerSensorService := createSensorService(home.ID, "sensor_power", []string{"W"}, "power")
	services = append(services, powerSensorService)

	currentPrice, err := t.tibber.GetCurrentPrice(home.ID)
	if err != nil {
		log.Error("Cannot get prices from Tibber - ", err)
		return
	}

	priceSensorService := createSensorService(home.ID, "sensor_price", []string{currentPrice.Currency}, "price")
	services = append(services, priceSensorService)

	incReort := fimptype.ThingInclusionReport{
		Address:        home.ID,
		CommTechnology: "tibber",
		ProductName:    "Tibber Pulse",
		Groups:         []string{"ch_0"},
		Services:       services,
		Alias:          home.AppNickname,
		ProductId:      "HAN Solo",
		DeviceId:       home.MeteringPointData.ConsumptionEan,
	}

	msg := fimpgo.NewMessage(
		"evt.thing.inclusion_report", "tibber", "object",
		incReort, nil, nil, oldMsg,
	)
	if err := t.mqt.RespondToRequest(oldMsg, msg); err == nil {
		log.WithError(err).Error("Could not publish MQTT message")
	}
	log.Debug("Inclusion report sent")
}