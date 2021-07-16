# PATCHFILES
I came to the idea to create patchfiles, when I saw lots of config files people create.
 
Patchfiles implements various config scripts into one single bash file which you can run on freshly installed system.

Implemented are:
* sysctl.conf
* open files limit
* sshd secure server
* bfq scheduler
* tune initial window size %

System automatically builds revert.sh script, whcih can bu run in order to revert back changes.

## PATCH (INSTALL)
```bash <(curl -Ss https://raw.githubusercontent.com/dpanic/patchfiles/main/patch_main.sh)```

## REVERT (UNINSTALL)
```bash <(curl -Ss https://raw.githubusercontent.com/dpanic/patchfiles/main/revert_main.sh)```



## TODO:
* implement stats ✅
* implement after patch command ✅
* implement revert %
* implement github ci/cd % 
    * docker % 
    * hooks % 
    * generate output patchfiles.sh file on every push to main/dev % 