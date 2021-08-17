# PATCHFILES
[![Go Report Card](https://goreportcard.com/badge/github.com/dpanic/patchfiles)](https://goreportcard.com/report/github.com/dpanic/patchfiles)

I came to the idea to create patchfiles, when I saw lots of config files people create.
 
Patchfiles implements various config scripts into one single bash file which you can run on freshly installed system.

## IMPLEMENTED:
* sysctl.conf
* open files limit
* sshd secure server
* bfq scheduler
* tune initial window size

System automatically builds revert.sh script, whcih can bu run in order to revert back changes.

## BUILT AND TESTED ON
* Ubuntu 20.04

## PATCH (INSTALL)
Start as a root:
```
bash <(curl -L -s https://github.com/dpanic/patchfiles/releases/latest/download/patch.sh)
```

## REVERT (UNINSTALL)
Start as a root:
```
bash <(curl -L -Ss https://github.com/dpanic/patchfiles/releases/latest/download/revert.sh)
```



## TODO
* ~~implement stats~~ ✅
* ~~implement after patch command~~ ✅
* ~~refactor code to use Go Lang HTML templates~~ ✅
* ~~implement detection if patched, used in patch script~~ ✅ 
* ~~implement detection if not patched, used in revert script~~ ✅ 
* ~~implement revert~~ ✅ 
* ~~implement revert move .old to .current file if overwrite used in patching~~ ✅ 
* ~~implement categories (networking, performance, security, general ...)~~ ✅
* ~~implement patch by category~~ ✅
* ~~implement revert by category~~ ✅
* implement github ci/cd % 
    * docker % 
    * hooks % 
    * generate output patchfiles.sh file on every push to main/dev % 

References:
* https://github.com/ncipollo/release-action
