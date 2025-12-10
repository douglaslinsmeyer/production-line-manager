#include <HardwareSerial.h>     // Reference the ESP32 built-in serial port library
#include "WS_WIFI.h"
#include "WS_Bluetooth.h"
#include "WS_GPIO.h"
#include "WS_RS485.h"
#include "WS_CAN.h"
#include "WS_RTC.h"
#include "WS_GPIO.h"
#include "WS_DIN.h"
#include "WS_SD.h"
#include "WS_ETH.h"


uint32_t Simulated_time=0;      // Analog time counting

/********************************************************  Initializing  ********************************************************/
void setup() { 
  Flash_test(); 
  GPIO_Init();  // RGB . Buzzer GPIO
  I2C_Init();
  RTC_Init();// RTC
  RS485_Init();
  CAN_Init();
  SD_Init();
  WIFI_Init();// WIFI
  Bluetooth_Init();// Bluetooth
  ETH_Init();
  
  DIN_Init();                 // If you don't want to control the dout through DIN, change Dout_Immediate_Default to 0 in WS_DIN.h and re-burn the program
  Dout_Init();

  printf("Connect to the WIFI network named \"ESP32-S3-POE-ETH-8DI-8DO\" and access the Internet using the connected IP address!!!\r\n");
}

/**********************************************************  While  **********************************************************/
void loop() {

}