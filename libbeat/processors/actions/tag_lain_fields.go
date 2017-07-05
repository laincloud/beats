package actions

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/processors"
	api "github.com/laincloud/lainlet/api/v2"
	"github.com/laincloud/lainlet/client"
	"github.com/laincloud/lainlet/watcher/container"
)

type tagLainFieldsConfig struct {
	LainletAddress string `config:"lainlet_address"`
}

type tagLainFields struct {
	lainletAddress string
	hostName       string
	holder         *dataHolder
}

// dataHolder holds the data struct
type dataHolder struct {
	data map[string]container.Info
}

func init() {
	processors.RegisterPlugin("tag_lain_fields",
		configChecked(newTagLainFields, allowedFields("when", "lainlet_address"), requireFields("lainlet_address")))
}

func newTagLainFields(c common.Config) (processors.Processor, error) {
	config := tagLainFieldsConfig{}
	err := c.Unpack(&config)
	if err != nil {
		return nil, fmt.Errorf("fail to unpack the tag_lain_fields configuration: %s", err)
	}
	var hostName string
	if hostName, err = os.Hostname(); err != nil {
		return nil, err
	}
	t := tagLainFields{
		lainletAddress: config.LainletAddress,
		hostName:       hostName,
		holder: &dataHolder{
			data: make(map[string]container.Info),
		},
	}
	go t.updateContainerInfo()
	return t, nil
}

func (t tagLainFields) Run(event common.MapStr) (common.MapStr, error) {
	if exist, _ := event.HasKey("container_id"); !exist {
		return nil, nil
	}
	containerID, _ := event.GetValue("container_id")
	var containerIDStr string
	var ok bool
	if containerIDStr, ok = containerID.(string); !ok {
		return nil, nil
	}
	if containerInfo, exist := t.holder.data[containerIDStr]; exist && containerIDStr != "" {
		event.Put("app_name", containerInfo.AppName)
		event.Put("proc_name", containerInfo.ProcName)
		event.Put("instance_no", containerInfo.InstanceNo)
		event.Put("app_version", containerInfo.AppVersion)
		event.Delete("container_id")
		return event, nil
	}
	return nil, nil
}

func (t tagLainFields) updateContainerInfo() {
	lainletClient := client.New(t.lainletAddress)
	url := fmt.Sprintf("/v2/containers?nodename=%s", t.hostName)
	idRe := regexp.MustCompile(fmt.Sprintf("^.{%d}/([a-z0-9]{12})[a-z0-9]+$", len(t.hostName)))
	for {
		//ctx, _ := context.WithTimeout(context.Background(), time.Hour*3)
		ch, err := lainletClient.Watch(url, context.Background())
		if err != nil {
			logp.Err("Error to watch lainlet: ", err.Error())
		} else {
			// There is no need to use any lock, we sacrifice accuracy to increase the throughput
			for event := range ch {
				if event.Event == "init" || event.Event == "update" || event.Event == "delete" {
					newData := new(api.GeneralContainers)
					if err = newData.Decode(event.Data); err != nil {
						logp.Err("Decode lainlet data error", err.Error())
					} else {
						shortIDData := make(map[string]container.Info, len(newData.Data))
						for key, cInfo := range newData.Data {
							matches := idRe.FindStringSubmatch(key)
							if len(matches) == 2 {
								shortIDData[matches[1]] = cInfo
							}
						}
						if !reflect.DeepEqual(shortIDData, t.holder.data) {
							logp.Info("App data changed: %v", shortIDData)
							t.holder.data = shortIDData
						}
					}
				}
			}
		}
		time.Sleep(time.Second * 3)
	}
}

func (t tagLainFields) String() string {
	return "lainlet_address=" + t.lainletAddress
}
