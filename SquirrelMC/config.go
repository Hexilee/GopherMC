package main

type ConfigType struct {
	APPName string `default:"SquirrelMC"`
	Description string `default:"Squirrel Message Channel"`
	LogFile string `default:"./log.txt"`
	Hub struct{
		Tcp struct{
			Service string `default:"127.0.0.1:8001"`
			MaxBytes int `default:"1024"`
		}
		WebSocket struct{
			Service string `default:"127.0.0.1:8002"`
			MaxBytes int `default:"1024"`
		}
	}
	Client struct{
		Tcp struct{
			Service string `default:"127.0.0.1:9001"`
			MaxBytes int `default:"1024"`
		}
		WebSocket struct{
			Service string `default:"127.0.0.1:9002"`
			MaxBytes int `default:"1024"`
		}
	}
}

var Config = ConfigType{}