go-mbpool (experimental!)
================================================================================

Very simple modbusRTU holding register reader with some convenient options like influxDB output. 
The tool has been created to READ remotely SUN2000KTL33 inverter by using raspberry Pi. 
The inverter supports rs485(modbusRTU). 

# Use cases

Imagine you have a solar panels+inverter that supports modbusRTU protocol like SUN2000KTL33. You've installed
the system in a remote village and managed to provide internet connectivity, but then you realize that by default
the inverter has no TCP/IP module, etc. You really want to check the metrics of your solar system remotely and keep a history of the values(possibly every 5mins).
This is where go-mbpool can help(probably and many others). Just get a raspberryPI, USBtoRS485 adapter, setup influxDB+grafana instance and add some
dashboards. The end result might look like this:

![Solar dashboard](/grafana.png)

# How to compile for raspberry Pi:

```
#PI3B+ ARM7
make pi

#PIZERO ARM6
make pi6
```

# How to use:

```
./go-mbpool.arm -interval 60s -rfile rfile -influxdb http://influx.domain.net:8086/write?db=dbname&u=v&p=p
```

Once started the reading will be printed to:

1. STDOUT

```
2020-07-04 01:50:23: Starting Solarmon ...

 active_power**|        0kW |32290| 20.4ms
  input_power**|        0kW |32294| 17.6ms
        ehour**|       0kWh |32298| 17.5ms
        ehour^p|       0kWh |32345| 17.6ms
         eday**|       0kWh |32300| 17.6ms
         eday^p|  223.42kWh |32349| 17.6ms
       emonth**|  685.17kWh |32302| 17.6ms
       emonth^p| 5343.55kWh |32353| 17.6ms
          eyear| 27025.69kWh |32304| 17.6ms
        eyear^p| 17602.93kWh |32357| 17.6ms
       etotal**| 44630.33kWh |32306| 17.6ms
         temp**|      37.1C |32286| 16.4ms
           Upv1|         0V |32262| 16.4ms
           Ipv1|         0A |32263| 16.4ms
           Upv2|         0V |32264| 16.4ms
           Ipv2|         0A |32265| 16.4ms
           Upv3|         0V |32266| 16.4ms
           Ipv3|         0A |32267| 16.4ms
           Upv4|         0V |32268| 16.3ms
           Ipv4|         0A |32269| 16.4ms
           Upv5|         0V |32270| 16.4ms
           Ipv5|         0A |32271| 16.4ms
           Upv6|         0V |32272| 17.6ms
           Ipv6|         0A |32273| 16.4ms
           Upv7|         0V |32314| 17.6ms
           Ipv7|         0A |32315| 16.4ms
           Upv8|         0V |32316| 16.4ms
           Ipv8|         0A |32317| 16.4ms
     efficiency|        0%% |32285| 16.4ms
       ongrid**|          0 |32322| 16.4ms
    iResistance|  3.266Mohm |32323| 16.4ms
      inP_MPPT1|        0kW |33022| 17.6ms
      inP_MPPT2|        0kW |33024| 17.6ms
      inP_MPPT3|        0kW |33026| 17.6ms
      inP_MPPT4|        0kW |33070| 17.6ms
            Uab|         0V |32274| 16.4ms
            Ubc|         0V |32275| 16.3ms
            Uca|         0V |32276| 16.4ms
             Ua|         0V |32277| 16.4ms
             Ub|         0V |32278| 16.3ms
             Uc|         0V |32279| 16.4ms
             Ia|         0A |32280| 16.4ms
             Ib|         0A |32281| 16.4ms
             Ic|         0A |32282| 16.4ms
           freq|        0Hz |32283| 16.4ms
   power_factor|          0 |32284| 16.4ms
     inv_status|      40960 |32287| 16.3ms
   peak_power**|        0kW |32288| 17.6ms
2020-07-04 01:50:23 ## 811.304218ms, 90.312µs; **


```


2. There is also HTTP preview:

```
curl http://loclhost:8090

or

curl http://loclhost:8090?long=1

```

3. If influxdb option is set the readings will be inserted there


# Options

```
Usage of /home/pi/solarpooler/go-mbpool.arm:
  -HTTPListen string
      HTTP listen addr (default ":8090")
  -br int
      Baud Rate (default 19200)
  -db int
      Data Bits (default 8)
  -influxDry
      Just print influx queries on stdout
  -influxTags string
      Tags to write with every measurement (default "loc=1,type=1,inverter=ktl33")
  -influxdb string
      Influx db: http://localhost:8086/write?db=dbmae&u=user&p=pass
  -interval duration
      Seconds to wait between reads (default 10s)
  -nightmode
      Sleep during the night
  -nightmodeEnd int
      Night ends at (default 5)
  -nightmodeSleep duration
      Check for night inerval (default 5m0s)
  -nightmodeStart int
      Night starts at (default 22)
  -once
      Run only once and exit
  -prty string
      Parity (default "N")
  -rfile string
      File with registers list
  -sb int
      Stop Bits (default 1)
  -slaveId uint
      Slave ID (default 15)
  -t duration
      Max secs to wait for single read to finish (default 5s)
  -tty string
      TTY device file (default "/dev/ttyUSB0")
  -v  show version

```

# Registers description file (rfile)
```
#register_id:read_count:gain:type:unit:timestamp offset:influxdb measurement name 
#active power
32290:2:active_power**:1000:I32:kW
#total input power
32294:2:input_power**:1000:U32:kW
#E-hour
32298:2:ehour**:100:U32:kWh:5m:ehour
32345:2:ephour:100:U32:kWh:-1h:ephour
#E-day
32300:2:eday**:100:U32:kWh:1h:eday
32349:2:epday:100:U32:kWh:-1d:epday
#E-Month
32302:2:emonth**:100:U32:kWh:1d:emonth
32353:2:epmonth:100:U32:kWh:-1m:epmonth
#E-Year
32304:2:eyear:100:U32:kWh:1d:eyear
#E-Total
32306:2:etotal**:100:U32:kWh:1d:etotal
#Cabinet temp
32286:1:temp**:10:I16:C
```

# TODO
* Use json for registers desc.
* Read alarms
* Add tests
* Cleanup

