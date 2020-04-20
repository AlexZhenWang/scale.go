package scalecodec

import (
	"encoding/json"
	scaleType "github.com/freehere107/scalecodec/types"
	"github.com/freehere107/scalecodec/utiles"
	"reflect"
)

type EventsDecoder struct {
	scaleType.Vec
	Metadata scaleType.MetadataCallAndEvent `json:"metadata"`
	Elements []map[string]interface{}       `json:"elements"`
}

func (e *EventsDecoder) Init(data scaleType.ScaleBytes, args []string) {
	e.TypeString = "Vec<EventRecord>"
	var metadata scaleType.MetadataCallAndEvent
	var subType string
	if len(args) > 0 {
		subType = args[0]
	}
	if len(args) > 1 {
		_ = json.Unmarshal([]byte(args[1]), &metadata)
	}
	e.Metadata = metadata
	e.ScaleDecoder.Init(data, &scaleType.ScaleDecoderOption{SubType: subType})
}

func (e *EventsDecoder) Process() []map[string]interface{} {
	elementCount := int(e.ProcessAndUpdateData("Compact<u32>").(int))
	bm, _ := json.Marshal(e.Metadata)
	er := EventRecord{}
	er.Init(e.Data, []string{"", string(bm)})
	for i := 0; i < elementCount; i++ {
		element := er.Process()
		element["event_idx"] = i
		e.Elements = append(e.Elements, element)
	}
	return e.Elements
}

type EventRecord struct {
	scaleType.ScaleDecoder
	Metadata     MetadataDecoder           `json:"metadata"`
	Phase        int                       `json:"phase"`
	ExtrinsicIdx int                       `json:"extrinsic_idx"`
	Type         string                    `json:"type"`
	Params       []map[string]interface{}  `json:"params"`
	Event        scaleType.MetadataEvents  `json:"event"`
	EventModule  scaleType.MetadataModules `json:"event_module"`
	Topic        []string                  `json:"topic"`
}

func (e *EventRecord) Init(data scaleType.ScaleBytes, args []string) {
	var metadata MetadataDecoder
	var subType string
	if len(args) > 0 {
		subType = args[0]
	}
	if len(args) > 1 {
		_ = json.Unmarshal([]byte(args[1]), &metadata)
	}
	e.Metadata = metadata
	e.ScaleDecoder.Init(data, &scaleType.ScaleDecoderOption{SubType: subType})
}

func (e *EventRecord) Process() map[string]interface{} {
	e.Phase = e.GetNextU8()
	if e.Phase == 0 {
		e.ExtrinsicIdx = int(e.ProcessAndUpdateData("U32").(uint))
	}
	e.Type = utiles.BytesToHex(e.NextBytes(2))
	// if e.MetadataV6.MetadataV6.EventIndex[e.Type] != nil {
	// 	eventIndex := e.MetadataV6.MetadataV6.EventIndex[e.Type].(map[string]interface{})
	// 	bc, _ := json.Marshal(eventIndex["call"])
	// 	var event scaleType.MetadataEvents
	// 	_ = json.Unmarshal(bc, &event)
	// 	e.Event = event
	// 	var EventModule scaleType.MetadataModules
	// 	bc, _ = json.Marshal(eventIndex["module"])
	// 	_ = json.Unmarshal(bc, &EventModule)
	// 	e.EventModule = EventModule
	// }
	for _, argType := range e.Event.Args {
		argTypeObj := e.ProcessAndUpdateData(argType)
		e.Params = append(e.Params, map[string]interface{}{
			"type":     argType,
			"value":    argTypeObj,
			"valueRaw": "",
		})
	}
	if utiles.StringToInt(e.Metadata.Version) >= 5 {
		topicValue := e.ProcessAndUpdateData("Vec<Hash>").([]interface{})
		for _, v := range topicValue {
			e.Topic = append(e.Topic, v.(reflect.Value).String())
		}
	}
	return map[string]interface{}{
		"phase":         e.Phase,
		"extrinsic_idx": e.ExtrinsicIdx,
		"type":          e.Type,
		"module_id":     e.EventModule.Name,
		"event_id":      e.Event.Name,
		"params":        e.Params,
	}

}
