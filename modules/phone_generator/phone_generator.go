package phone_generator

import (
	"fmt"
	"github.com/graniet/operative-framework/session"
	"os"
	"strings"
	"sync"
	"syreclabs.com/go/faker"
	"gopkg.in/cheggaaa/pb.v1"
	"github.com/segmentio/ksuid"
)

type PhoneGenerator struct{
	session.SessionModule
	Sess *session.Session
	Phones []string
	Current int
	Bar *pb.ProgressBar
}

func PushPhoneGeneratorModule(s *session.Session) *PhoneGenerator{
	mod := PhoneGenerator{
		Sess: s,
		Phones: []string{},
		Current: 1,
	}

	mod.CreateNewParam("NUMBER_PREFIX", "Country prefix ex: (310)", "310", false, session.STRING)
	mod.CreateNewParam("NAME_PREFIX", "Prefix of contact random name ex: (BHILLS_)", "", false, session.STRING)
	mod.CreateNewParam("FILE_PATH", "Location for generated VCards", "", false, session.STRING)
	mod.CreateNewParam("VCARD", "Generate vcard to file", "true", false, session.STRING)
	return &mod
}

func (module *PhoneGenerator) Name() string{
	return "phone_generator"
}

func (module *PhoneGenerator) Description() string{
	return "Generate VCard (.vcf) with random number for whatsapp OSINT"
}

func (module *PhoneGenerator) Author() string{
	return "Tristan Granier"
}

func (module *PhoneGenerator) GetType() string{
	return "country"
}

func (module *PhoneGenerator) GetInformation() session.ModuleInformation{
	information := session.ModuleInformation{
		Name: module.Name(),
		Description: module.Description(),
		Author: module.Author(),
		Type: module.GetType(),
		Parameters: module.Parameters,
	}
	return information
}

func (module *PhoneGenerator) Start(){

	argumentPrefix, err := module.GetParameter("NUMBER_PREFIX")
	if err != nil{
		argumentPrefix = session.Param{
			Value: "",
		}
	}
	argumentFilePath, err2 :=  module.GetParameter("FILE_PATH")
	if err2 != nil{
		argumentFilePath = session.Param{
			Value: "",
		}
	}
	argumentNamePrefix, err3 := module.GetParameter("NAME_PREFIX")
	if err3 != nil{
		argumentNamePrefix = session.Param{
			Value: "",
		}
	}

	argumentVCard, err4 := module.GetParameter("VCARD")
	if err4 != nil{
		argumentVCard = session.Param{
			Value: "",
		}
	}


	module.Bar = pb.New(5000)

	pool, err := pb.StartPool(module.Bar)
	if err != nil {
		panic(err)
	}
	wg := new(sync.WaitGroup)
	for{
		if module.Current < 5000 {
			wg.Add(1)
			go func(module *PhoneGenerator, bar *pb.ProgressBar) {
				phone := faker.PhoneNumber().CellPhone()
				if strings.Contains(phone, "(") && strings.Contains(phone, ")") {
					newPhone := strings.Split(phone, ")")[1]
					if argumentPrefix.Value != "" {
						newPhone = "+1 ("+strings.TrimSpace(argumentPrefix.Value)+")" + newPhone
					} else{
						newPhone = "+1 (310)" + newPhone
					}

					module.Phones = append(module.Phones, newPhone)
					module.Current = module.Current + 1
					bar.Increment()
				}
			}(module, module.Bar)
			wg.Done()
		} else{
			break
		}
	}
	wg.Wait()
	_ = pool.Stop()

	var file *os.File
	var errPath error

	if argumentFilePath.Value != ""{
		file, errPath = os.OpenFile(strings.TrimSpace(argumentFilePath.Value), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	} else {
		file, errPath = os.OpenFile("/Users/graniet/Desktop/VCARD/beverlyHills-5000_1.vcf", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	}
	if errPath != nil {
		fmt.Println(errPath.Error())
		return
	}
	defer file.Close()
	for _, number := range module.Phones{
		var uuid string
		if argumentNamePrefix.Value == "" {
			uuid = "BHills_GO_" + ksuid.New().String()
		} else{
			uuid = strings.TrimSpace(argumentNamePrefix.Value) + "_" + ksuid.New().String()
		}
		if argumentVCard.Value == "true" {
			_, _ = file.WriteString("BEGIN:VCARD\nVERSION:3.0\nN:" + uuid + ";;;\nFN:" + uuid + "\nTEL;type=HOME:" + number + "\nEND:VCARD\n")
		} else{
			_, _ = file.WriteString("\"" + number + "\",\n")
		}
	}
	module.Sess.Stream.Success("VCards successfully generated")
}