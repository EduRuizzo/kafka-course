##########
# SECURITY
##########

# Generate the private key
keytool -keystore server.keystore.jks -alias localhost -validity 365 -genkey -keyalg RSA

# You can view the contents of the keystore at any time — by running the following command.
keytool -list -v -keystore server.keystore.jks

# Create a CA
openssl req -new -x509 -keyout ca-key -out ca-cert -days 365

# The next two steps will import the resulting ca-cert file to the broker and client truststores. Once
# imported, the parties will implicitly trust the CA and any certificate signed by the CA.
keytool -keystore client.truststore.jks -alias CARoot -import -file ca-cert
keytool -keystore server.truststore.jks -alias CARoot -import -file ca-cert

# Sign the broker certificate.The next step is to generate the certificate signing request on behalf of the broker.
keytool -keystore server.keystore.jks -alias localhost -certreq -file cert-req

# This produces cert-req, being the signing request. To sign with the CA, run the following command.
openssl x509 -req -CA ca-cert -CAkey ca-key -in cert-req -out cert-signed -days 365 -CAcreateserial

# The CA certificate must be imported into the server’s keystore under the CARoot alias.
keytool -keystore server.keystore.jks -alias CARoot -import -file ca-cert

# Then, import the signed certificate into the server’s keystore under the localhost alias.
keytool -keystore server.keystore.jks -alias localhost -import -file cert-signed

# next step is to install the private key and the signed certificate on the broker. Assuming the keystore file is in /tmp/kafka-ssl, run:
cp /tmp/kafka-ssl/server.*.jks $KAFKA_HOME/config

# Generate a private key for the client application, using localhost as the value for the ‘first and last name attribute, leaving all others blank.
keytool -keystore client.keystore.jks -alias localhost -validity 365 -genkey -keyalg RSA

# Sign the client certificate. This produces the client certificate signing request file — client-cert-req.
keytool -keystore client.keystore.jks -alias localhost -certreq -file client-cert-req

# Next, we will action the signing request with the existing CA:
openssl x509 -req -CA ca-cert -CAkey ca-key -in client-cert-req -out client-cert-signed -days 365 -CAcreateserial

# The result is the client-cert-signed file, ready to be imported into the client’s keystore, along with the CA certificate.
# Starting with the CA certificate:
keytool -keystore client.keystore.jks -alias CARoot -import -file ca-cert

# Moving on to the signed certificate:
keytool -keystore client.keystore.jks -alias localhost -import -file client-cert-signed

# Verify with CLI
kafka-topics.sh --list --bootstrap-server [::1]:9093 --command-config $KAFKA_HOME/config/ssl.config

# IMPORT IN GO
#
# The first steps to easily handle your certificates from Go is to convert them to a set of PEM files.
# Here are the commands to extract the Certificate Authority (CA) certificate:
keytool -importkeystore -srckeystore server.truststore.jks -destkeystore server.p12 -deststoretype PKCS12
openssl pkcs12 -in server.p12 -nokeys -out server.cer.pem

# You can then convert your client keystore to be usable from Go, with similar commands:
keytool -importkeystore -srckeystore server.keystore.jks -destkeystore client.p12 -deststoretype PKCS12
openssl pkcs12 -in client.p12 -nokeys -out client.cer.pem
openssl pkcs12 -in client.p12 -nodes -nocerts -out client.key.pem