#pragma once

#include <ArduinoJson.h>
#include <Arduino.h>
#include <PubSubClient.h>
#include <WiFi.h>
#include <WiFiClientSecure.h>
#include "WS_GPIO.h"
#include "WS_Information.h"
#include "WS_Dout.h"
#include "WS_WIFI.h"
#include "WS_Struct.h"

#define MSG_BUFFER_SIZE (50)


void MQTT_Init(void);
void MQTTTask(void *parameter);

void callback(char* topic, byte* payload, unsigned int length);   // MQTT subscribes to callback functions for processing received messages
void reconnect(void);                                                 // Reconnect to the MQTT server
void sendJsonData(void);                                              // Send data in JSON format to MQTT server


