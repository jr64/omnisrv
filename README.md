# Omnisrv

This repository contains **unofficial** and **unsupported** code and instructions for modding Tolino Shine 3 devices with Firmware 15.2.0 (the newest at the time of writing).

The modifications performed are potentially dangerous and **can** brick your device if you do not know what you are doing. Proceed with caution and **AT YOUR OWN RISK**.

## Overview

The instructions presented here are meant to document the modifications I made to my Tolino Shine 3. They are not intended for non-technical users but rather for anyone interested in the technical aspects.

I started out with the modifications documented at https://allesebook.de/anleitung/tolino-shine-3-adb-root-und-eigene-apps-installieren-video-978782/ (text is in german) and then added my own improvements to suit my needs.

* Faster turn-on time for the screen (1 second faster)
* To read eBooks, I prefer KOReader as it renders some more complicated layouts better for me
* I use RelaunchX rather than Smart Launcher, I find it easier to use with eInk
* I wanted a faster way to switch from the Tolino app to the Android launcher as I did not like OneToucher

The mod consists of two steps:

1. Patching android.policy.jar to significantly speed up turning the screen on
2. Installing the omnisrv executable to
    * allow turning the screen on by double tapping it anywhere (This does not drain battery, the functionality was already there in the hardware and also turned on. I simply added code to handle the touch events.)
    * launch ReLaunchX with a two finger swipe from top to bottom
    * turn off the screen with a two finger swipe from bottom to top
    * modify permissions to the backlight so it can be controlled from KOReader

# Build

In order to build the binary for the architecture of the Tolino Shine 3 you need a cross-compilation toolchain with musl. I recommend arm-linux-musleabihf-cross from https://musl.cc/

```bash
CC=/path/to/arm-linux-musleabihf-cross/bin/arm-linux-musleabihf-gcc CGO_ENABLED=1 GOOS="linux" GOARCH="arm" go build -ldflags="-s -w -extldflags=-static"
```

# Patching android.policy.jar

To remove the ~1 second white screen displayed every time the screen is turned on, you have to modify the android.policy.jar file. Please note that this is a core system library. If you make any mistakes, chances are high your device will not turn on anymore.

To perform the steps, you need ADB root access to your device, see below for details on how to create a modified boot image.

## Pull android-policy.jar from device
```bash
adb pull /system/framework/android-policy.jar
```

## Decompile jar
```bash
apktool d android.policy.jar
```

## Patch sleep duration from 1000000ns to 0ns
```bash
vim android.policy.jar.out/smali/ntx/power/PowerManagerEx.smali

on line 233:
old: const v4, 0xf4240
new: const v4, 0x0 
```

## Rebuild jar
```bash
apktool b -o mod.jar android.policy.jar.out
unzip mod.jar
jar -uf android.policy.jar -C . classes.dex # because apktool creates mod.jar without META-INF
mv android-policy.jar modded-android-policy.jar
```

## Copy jar to device
```bash
adb shell mkdir /mnt/media_rw/sdcard1/android-policy-mod
adb push modded-android-policy.jar /mnt/media_rw/sdcard1/android-policy-mod
```

## Remount /system as rw
```bash
adb shell mount -o remount,rw /system
```

## Start shell on device

```bash
adb shell
```
The next steps after this are performed directly on the device.

## Backup old jar and odex
```bash
cp /system/framework/android-policy.jar /mnt/media_rw/sdcard1/android-policy-mod/android.policy.jar.back
cp /system/framework/android-policy.odex /mnt/media_rw/sdcard1/android-policy-mod/android.policy.odex.back

cd /mnt/media_rw/sdcard1/android-policy-mod/
```

## Create odex
```bash
dexopt --preopt modded-android-policy.jar modded-android-policy.odex  m=y,u=y
# spoof old odex signature to avoid loading failures
busybox dd if=/system/framework/android.policy.odex of=modded-android-policy.odex bs=1 count=20 skip=52 seek=42 conv=notrunc
```

## Backup & delete dalvik cache
```bash
cp /data/dalvik-cache/system@framework@android.policy.jar@classes.dex .
mv system@framework@android.policy.jar@classes.dex system@framework@android.policy.jar@classes.dex.back
rm /data/dalvik-cache/system@framework@android.policy.jar@classes.dex
```

## Overwrite system files with modded jar and odex 
```bash
cp modded-android-policy.jar /system/framework/android.policy.jar
cp modded-android-policy.odex /system/framework/android.policy.odex
```

## Force reload of jar/odex
```bash
ps | grep servicemanager
kill <pid of servicemanager>
```

Finally, wait until boot animation is done, then fully reboot system if it isn't broken. If it doesn't work anymore: good luck.

# Install omnisrv

In order to get the omnisrv executable to run on boot, adb push it to /data/ (remember to chmod +x it) and then create a modified boot image that runs the executable on boot through init.rc

Please note that you will ALWAYS have to boot the Tolino with this modified image through fastboot if you want the executable to run.

Also rember to use ADB to install KOReader and RelaunchX.

## Download firmware update
Get update.zip from https://mytolino.de/software-updates-tolino-ereader/ and extract boot.img

## Extract boot.img
```bash
abootimg -x boot.img
mkdir ramdisk
cd ramdisk
gunzip -c ../initrd.img | cpio -i
```

## Modify init process

Add the following line after the other imports at the top of `init.rc`
```
import /init.omnisrv.rc
```

Create `init.omnisrv.rc` with the following content:
```
on boot
    # allow everyone to modify backlight so KOReader can change it
    chmod 666 /sys/class/backlight/lm3630a_led/color
service omnisrv /data/omnisrv
    class main
```

then:
```
chmod 750 init.omnisrv.rc
```

## OPTIONAL steps if you want to get a root shell via ADB
```
vim default.prop
```

```
persist.service.adb.enable=1
persist.sys.usb.config=mass_storage,adb
```

You will need a modded adb version that does not drop root privileges when opening a shell, you can find that on the internet.

```
cp /path/to/adbd_patched sbin/adbd
vim default.prop
```

```
ro.secure=0
ro.debuggable=1
```


## Repack boot.img
```
find . | cpio -o -H newc | gzip > ../newramdisk.img
cd ..
cat bootimg.cfg | grep -v "^bootsize" > bootimg-modified.cfg
abootimg --create boot-modified.img -k zImage -r newramdisk.img -f bootimg-modified.cfg
```

## Check modifications
```
abootimg -i boot.img
abootimg -i boot-modified.img 
```
(new id will be all zeroes but that should not matter, it is normally not checked)


## Boot tolino with modified image

```
fastboot boot boot-modified.img
```

1. Hold down power button until menu appears, choose "Switch off"
2. Wait until device is powered off
3. Hold down power button for ~30sec to enter fastboot

