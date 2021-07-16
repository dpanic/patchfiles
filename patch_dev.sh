
#!/bin/bash
#
# PATCHFILES
# 
# author: dpanic@gmail.com
# version: 0.0.1
# environment: dev
# built: 2021-07-16 12:15:51 +00:00
#
#

#
# command 'bfq_1'
#
# description:
#    load bfq modules
#
# body:
#    bfq
#
echo "Patching 'bfq_1'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCmJmcQojIFBBVENIRklMRVMgRU5ECg==" | base64 -d - > /etc/modules-load.d/bfq.conf

modprobe bfq

#
# command 'bfq_2'
#
# description:
#    load bfq modules
#
# body:
#    ACTION=="add|change", KERNEL=="sd*[!0-9]|sr*|nvme*|mmcblk*", ATTR{queue/scheduler}="bfq"
#    
#
echo "Patching 'bfq_2'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCkFDVElPTj09ImFkZHxjaGFuZ2UiLCBLRVJORUw9PSJzZCpbITAtOV18c3IqfG52bWUqfG1tY2JsayoiLCBBVFRSe3F1ZXVlL3NjaGVkdWxlcn09ImJmcSIKCiMgUEFUQ0hGSUxFUyBFTkQK" | base64 -d - > /etc/udev/rules.d/60-scheduler.rules

sudo udevadm control --reload
sudo udevadm trigger

#
# command 'limits_1'
#
# description:
#    implements limits for number of file descriptors (opened files), for all users as well as for user root. implements as well limits of number of processess started for all users as well as for user root.
#
# body:
#    *         hard    nofile      524288
#    *         soft    nofile      524288
#    root      hard    nofile      524288
#    root      soft    nofile      524288
#    
#    
#    *         soft    nproc       10240
#    *         hard    nproc       10240
#    root      soft    nproc       10240
#    root      hard    nproc       10240
#    
#    *         hard    stack       131072
#    *         soft    stack       131072
#    
#
echo "Patching 'limits_1'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCiogICAgICAgICBoYXJkICAgIG5vZmlsZSAgICAgIDUyNDI4OAoqICAgICAgICAgc29mdCAgICBub2ZpbGUgICAgICA1MjQyODgKcm9vdCAgICAgIGhhcmQgICAgbm9maWxlICAgICAgNTI0Mjg4CnJvb3QgICAgICBzb2Z0ICAgIG5vZmlsZSAgICAgIDUyNDI4OAoKCiogICAgICAgICBzb2Z0ICAgIG5wcm9jICAgICAgIDEwMjQwCiogICAgICAgICBoYXJkICAgIG5wcm9jICAgICAgIDEwMjQwCnJvb3QgICAgICBzb2Z0ICAgIG5wcm9jICAgICAgIDEwMjQwCnJvb3QgICAgICBoYXJkICAgIG5wcm9jICAgICAgIDEwMjQwCgoqICAgICAgICAgaGFyZCAgICBzdGFjayAgICAgICAxMzEwNzIKKiAgICAgICAgIHNvZnQgICAgc3RhY2sgICAgICAgMTMxMDcyCgojIFBBVENIRklMRVMgRU5ECg==" | base64 -d - > /etc/security/limits.conf


#
# command 'limits_2'
#
# description:
#    adds pam_limits kernel module to the pam in order to enable it for DESKTOP sessions
#
# body:
#    session required pam_limits.so
#    
#
echo "Patching 'limits_2'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCnNlc3Npb24gcmVxdWlyZWQgcGFtX2xpbWl0cy5zbwoKIyBQQVRDSEZJTEVTIEVORAo=" | base64 -d - >> /etc/pam.d/common-session


#
# command 'limits_3'
#
# description:
#    adds pam_limits kernel module to the pam in order to enable it for SSH sessions
#
# body:
#    session required pam_limits.so
#    
#
echo "Patching 'limits_3'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCnNlc3Npb24gcmVxdWlyZWQgcGFtX2xpbWl0cy5zbwoKIyBQQVRDSEZJTEVTIEVORAo=" | base64 -d - >> /etc/pam.d/common-session-noninteractive


#
# command 'limits_4'
#
# description:
#    for systems using systemd this append is needed to increase number of opened files
#
# body:
#    DefaultLimitNOFILE=1048576
#    
#
echo "Patching 'limits_4'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCkRlZmF1bHRMaW1pdE5PRklMRT0xMDQ4NTc2CgojIFBBVENIRklMRVMgRU5ECg==" | base64 -d - >> /etc/systemd/system.conf


#
# command 'limits_5'
#
# description:
#    for systems using systemd this append is needed to increase number of opened files
#
# body:
#    DefaultLimitNOFILE=1048576
#    
#
echo "Patching 'limits_5'"
echo "IyBQQVRDSEZJTEVTIFNUQVJUCkRlZmF1bHRMaW1pdE5PRklMRT0xMDQ4NTc2CgojIFBBVENIRklMRVMgRU5ECg==" | base64 -d - >> /etc/systemd/user.conf


