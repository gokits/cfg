# cfg: Reloadable config package for golang
## Features
1. Use golang struct instead of map for config.
1. Versioned and Atomic reloadable configuration
1. Customizable decoders(json out of box)
1. Customizable config source(reloadable file source out of box)
1. Hooks in config lifecycle, and can be used to customize behavior of cfg package:
   1. PreDecoder: called before new config decoding from source. It can be used to set default value.
   1. PostDecoder: called after new config decoding from source. It can be used to validate new config.
   1. PostSwap: called after new config taked effect(implemented by atomic swap pointers). It can be used to notify that new config is reloaded.
1. Simple logger interface to log events. 

## How to use
```golang
import (
    validator "github.com/go-playground/validator/v10"
    "github.com/gokits/cfg"
    "github.com/gokits/cfg/source/file"
)
//...

// 1. define config struct
// 2. define validator(optional)
type Config struct {
    MaxRetry               int  `validate:"min=0,max=10"`
    LogPath                string
}

var (
    gvalidator *validator.Validate
    filesource *file.File
    meta       *cfg.ConfigMeta
)
//...

// define PostDecode hook to validate new config. New config reload will be canceled if `error != nil`.
// `oldptr` will hold pointer to current version config struct
// `c` will hold pointer to new config struct
func (c *Config) PostDecode(oldptr interface{}) error {
    return gvalidator.Struct(c)
}

//...

func main() { 
    gvalidator = validator.New()
    if filesource, err = file.NewFileSource("./tmp.json"); err != nil {
        return
    }
    defer filesource.Close()
    meta = cfg.NewConfigMeta(Config{}, filesource)
    go meta.Run()
    if err = meta.WaitSynced(); err != nil {
        return
    }
    fmt.Printf("%+v\n", meta.Get().(*Config))
}
```

## Roadmap
1. [done] reloadable file source
1. [done] custimizable hooks
1. [todo] export config version
1. [todo] frozen package api and release v1
1. [todo] implement other config source(etcd, consul...)

## Status
This package has been used in production environment running on our kubernetes clusters. It works well with configmap of kubernetes and files. 
