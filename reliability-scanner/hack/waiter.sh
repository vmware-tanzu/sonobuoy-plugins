OUTPUT="make results"; 
while [ `echo $OUTPUT | grep -c status` -lt 0 ];
do
  echo "Waiting for Sonobuoy to complete......";
  sleep 10;
  OUTPUT=`$cmd`; 
done
