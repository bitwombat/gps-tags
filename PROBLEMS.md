Charlie is under the eave by our office, yet the GPS result is down by the
river:


2024/06/11 12:25:30
{"SerNo":810244,"IMEI":"353785725696453","ICCID":"89610180004127201907","ProdId":97,"FW":"97.2.1.11","Records":[{"SeqNo":34385,"Reason":1,"DateUTC":"2024-06-11 02:24:38","Fields":[{"GpsUTC":"2024-06-11 02:03:42","Lat":-31.4577534,"Long":152.6423117,"Alt":29,"Spd":0,"SpdAcc":7,"Head":165,"PDOP":19,"PosAcc":17,"GpsStat":3,"FType":0},{"DIn":2,"DOut":0,"DevStat":3,"FType":2},{"AnalogueData":{"1":4684,"3":2400,"4":14,"5":4594},"FType":6}]},{"SeqNo":34386,"Reason":6,"DateUTC":"2024-06-11 02:25:26","Fields":[{"GpsUTC":"2024-06-11 02:25:26","Lat":-31.4584367,"Long":152.6447267,"Alt":-464,"Spd":0,"SpdAcc":19,"Head":53,"PDOP":22,"PosAcc":33,"GpsStat":3,"FType":0},{"DIn":2,"DOut":0,"DevStat":3,"FType":2},{"AnalogueData":{"1":4684,"3":2400,"4":14,"5":4594},"FType":6}]}]}

Formatted a little:
{"SeqNo":34385,"Reason":1,"DateUTC":"2024-06-11 02:24:38","Fields":[{"GpsUTC":"2024-06-11 02:03:42","Lat":-31.4577534,"Long":152.6423117,"Alt":29,"Spd":0,"SpdAcc":7,"Head":165,"PDOP":19,"PosAcc":17,"GpsStat":3,"FType":0},{"DIn":2,"DOut":0,"DevStat":3,"FType":2},{"AnalogueData":{"1":4684,"3":2400,"4":14,"5":4594},"FType":6}]},
{"SeqNo":34386,"Reason":6,"DateUTC":"2024-06-11 02:25:26","Fields":[{"GpsUTC":"2024-06-11 02:25:26","Lat":-31.4584367,"Long":152.6447267,"Alt":-464,"Spd":0,"SpdAcc":19,"Head":53,"PDOP":22,"PosAcc":33,"GpsStat":3,"FType":0},{"DIn":2,"DOut":0,"DevStat":3,"FType":2},{"AnalogueData":{"1":4684,"3":2400,"4":14,"5":4594},"FType":6}]}]}

Reason 1 is "start of trip"
Reason 6 is "distance travelled"

Note that this tag is in the "trip" mode, not periodic, 10-minute reports like
Rueger's.

Note the PosAcc of 33. Need to compare that to others.
Maybe it needs to be thresholded.
