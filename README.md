# dlayer
Stats collector for Docker layers

Downloads automatically generated by wercker:

  * Linux: https://s3.amazonaws.com/wercker-development/dlayer/master/latest/linux_amd64/dlayer
  * OS X: https://s3.amazonaws.com/wercker-development/dlayer/master/latest/darwin_amd64/dlayer
  
Looks something like:
```
termie@champs-3:dev/wercker/dlayer % ./dlayer sizes
Tag nodesource/trusty:latest                :   12 layers -  356MB (virtual)
Tag mongo:latest                            :   17 layers -  244MB (virtual)
Tag tcnksm/gox:1.4.1                        :   13 layers - 1702MB (virtual)
Tag tcnksm/gox:latest                       :   13 layers - 1704MB (virtual)
Tag ubuntu:latest                           :    4 layers -  179MB (virtual)
Tag nodesource/node:trusty                  :   12 layers -  386MB (virtual)
Tag google/python:latest                    :    6 layers -  362MB (virtual)
Tag tcnksm/gox:1.4.2                        :   13 layers - 1704MB (virtual)
Tag google/golang:latest                    :    9 layers -  583MB (virtual)
Tag redis:latest                            :   17 layers -  104MB (virtual)
Tag golang:latest                           :   14 layers -  493MB (virtual)
Tag nginx:latest                            :   12 layers -  126MB (virtual)
Tag busybox:latest                          :    3 layers -    2MB (virtual)
Tag phusion/passenger-ruby22:latest         :   16 layers -  635MB (virtual)
Total    :  163 layers -   8832MB (actual)
Reachable:  149 layers -   8339MB (actual)
                           8585MB (virtual)
Shared   :    5 layers -    165MB (actual)
```
