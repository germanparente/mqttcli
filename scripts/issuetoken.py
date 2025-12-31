import sys
import json
import requests
user="cgermanparente@gmail.com"
if len(sys.argv) <= 1:
	print("Password missing\n")
	exit(1)
password=sys.argv[1]
envoy_serial="122318042179"
data = {'user[email]': user, 'user[password]': password}
response = requests.post("http://enlighten.enphaseenergy.com/login/login.json?", data=data)
response_data = json.loads(response.text)
data = {'session_id': response_data['session_id'], 'serial_num': envoy_serial, 'username': user}
response = requests.post('http://entrez.enphaseenergy.com/tokens', json=data)
token_raw = response.text
print(token_raw)
