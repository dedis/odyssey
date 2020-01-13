#!/bin/sh

# In this variant we use a cothority hosted on a remote server. We take the
# assumption that the ledger has already been created and informations are saved
# in the 'secret/' folder.

#
# Build conode, bcadmin, and csadmin
#

# Key are in the form of (key || initVal)
# where the key has 32 hex char and the initVal 24
KEY1="b5de94468a74b4586ac4acab12fb409bf9d61f725b331fce8778d91e"
KEY2="4d0479cb26816af63f9fc997c3bcbf7e50baf4992e18e925c6e4eb72"

export BC_CONFIG="/Users/nkocher/GitHub/odyssey/secret"
CSA="csadmin"
BCA="bcadmin"

#
# Create roster
#

OUTRES=$($BCA create $BC_CONFIG/roster.toml)
bcPATH=$(echo $OUTRES | grep BC= | cut -d '"' -f 2)
bcID=$(echo "$bcPATH" | sed -e "s/.*bc-\(.*\).cfg/\1/" )
export BC="$bcPATH"

#
# Authorize conodes
#

# THIS STEP MUST BE DONE ON THE SEVER WITH
# `csadmin a co1/private.toml BYZCOIN_ID`

#
# Create DARC
#

$BCA darc add -out_id $BC_CONFIG/darc_id.txt -out_key $BC_CONFIG/darc_key.txt -unrestricted

ID=`cat $BC_CONFIG/darc_id.txt`
KEY=`cat $BC_CONFIG/darc_key.txt`

# WARNING sometime a counter error can occur if all those transactions are sent
# in once. Try re-submitting them if something goes wrong.
$BCA darc rule -rule "spawn:longTermSecret" --darc $ID --sign $KEY --identity $KEY
$BCA darc rule -rule "spawn:calypsoWrite" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "spawn:calypsoRead" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "spawn:value" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:value.update" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "spawn:odysseyproject" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:odysseyproject.update" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:odysseyproject.updateStatus" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:odysseyproject.setURL" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:odysseyproject.setAccessPubKey" -darc $ID -sign $KEY -identity $KEY
$BCA darc rule -rule "invoke:odysseyproject.setEnclavePubKey" -darc $ID -sign $KEY -identity $KEY
bcadmin darc rule -rule spawn:darc -darc darc:789d38f8f2ed19006fa58c38491f80312143afee791ad9bbf812f77b903fcd2f -id ed25519:32804067da1f1ed9fb465dd4249bc685a849b28b4e2b4842ea39af1f752f71d1 -sign ed25519:32804067da1f1ed9fb465dd4249bc685a849b28b4e2b4842ea39af1f752f71d1
# attr:allowed:uses=clientData&purposes=clientData,financeData
#
# Spawn LTS
#

LTS_ID=$($CSA contract lts spawn --darc "$ID" --sign "$KEY" -x)
# I don't trust myself
if ! [[ $LTS_ID =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong LTS_ID: '$LTS_ID'"
fi

# 
# Create LTS and save the public key
#

PUB_KEY=$($CSA dkg start --instid "$LTS_ID" -x)

if ! [[ $PUB_KEY =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong PUB_KEY: '$PUB_KEY'"
fi

#
# Spawn write 1
#
read -r -d '' DATA << EOM
{
	"Title": "Titanic",
	"Description": "The Titanic dataset",
	"CloudURL": "dedis/datasets/1_titanic.csv.aes",
	"Author": "John Doe",
	"WriteInstID": "",
	"SHA2": "7d118fef8b6ccf7f81111877bc388536f7b1e498a655e3d649d19aaa010e9f6f"
}
EOM

WRITE_ID=$($CSA contract write spawn -x --darc "$ID" --extraData "$DATA" --sign "$KEY" --instid "$LTS_ID" --secret "$KEY1" --key "$PUB_KEY")

if ! [[ $WRITE_ID =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong WRITE_ID: '$WRITE_ID'"
fi

#
# Save the write instance ids in a value contract, separated by comas
#

OUTRES=$($BCA contract value spawn --value "$WRITE_ID" --darc "$ID" --sign "$KEY")
VALUE_INSTANCE_ID=$( echo "$OUTRES" | grep -A 1 "instance id" | sed -n 2p )

if ! [[ $VALUE_INSTANCE_ID =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong VALUE_INSTANCE_ID: '$VALUE_INSTANCE_ID'"
fi

#
# Spawn write 2
#
read -r -d '' DATA << EOM
{
	"Title": "Movies actors",
	"Description": "Images of celebrities from various movies",
	"CloudURL": "dedis/datasets/2_images.tar.aes",
	"Author": "Michel Fritz",
	"WriteInstID": "",
	"SHA2": "bdf7daecaf76d97ef0fed0fe3230a852ddb1814cf0ad20680f382ec29b6ca5e5"
}
EOM

WRITE_ID=$($CSA contract write spawn -x --darc "$ID" --extraData "$DATA" --sign "$KEY" --instid "$LTS_ID" --secret "$KEY2" --key "$PUB_KEY")

if ! [[ $WRITE_ID =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong WRITE_ID: '$WRITE_ID'"
fi

#
# Update the value contract
#

CONTENT=$($BCA contract value get -i $VALUE_INSTANCE_ID)
$BCA contract value invoke update -i $VALUE_INSTANCE_ID --value "$CONTENT,$WRITE_ID" --darc "$ID" --sign "$KEY"

#
# Print the instance id and the list stored in the value contract
#

echo "Instance id:"
echo "$VALUE_INSTANCE_ID"

CONTENT=$($BCA contract value get -i $VALUE_INSTANCE_ID)
echo "The content is:"
echo "$CONTENT"

#
# Iterate over the instance ids saved
#

# This is not an optimized way, as we are cutting each time
i=1
j="$(echo "$CONTENT" | cut -d "," -f $i)" 
while [ "$j" ]
do
    echo "Instance id $i: $j"
    i="$(($i+1))"
    j="$(echo "$CONTENT" | cut -d "," -f $i)" 
done 


#
# Spawn read
#

OUTRES=$($CSA contract read spawn --sign $KEY --instid $WRITE_ID)
READ_ID=`echo "$OUTRES" | sed -n '2p'` # must be at the second line

if ! [[ $READ_ID =~ ^[0-9a-f]{64}$ ]]; then
    echo "wrong READ_ID: '$READ_ID'"
fi

#
# Creation of the DARC
#

# from the "secret" folder
# Data owner 1
bcadmin -c . darc add --out_id data_owner_1/darc_id.txt --out_key data_owner_1/darc_key.txt --desc "DARC of the data owner N°1"
# Darc for the titanic Dataset
bcadmin -c . darc add --out_id data_owner_1/titanic/darc_id.txt --out_key data_owner_1/titanic/darc_key.txt --desc "DARC for the titanic dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_1/titanic/darc_id.txt) --id $(cat data_owner_1/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_1/titanic/darc_id.txt) --id $(cat data_owner_1/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_1/titanic/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_1/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 1 should be able to update the darc for the titanic dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_1/titanic/darc_id.txt) --id "$(cat data_owner_1/darc_id.txt) | attr:allowed:uses=clientData&purposes=clientData,financeData" --sign $(cat data_owner_1/darc_key.txt) --replace --restricted

# Add the titanic dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Titanic",
	"Description": "The Titanic dataset",
	"CloudURL": "dedis/datasets/1_titanic.csv.aes",
	"Author": "John Doe",
	"WriteInstID": "",
	"SHA2": "7d118fef8b6ccf7f81111877bc388536f7b1e498a655e3d649d19aaa010e9f6f"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_1/titanic/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_1/darc_key.txt) --instid $(cat lts_id.txt) --secret b5de94468a74b4586ac4acab12fb409bf9d61f725b331fce8778d91e --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_1/titanic/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_1/titanic/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)

 # Darc for the MOVIE DATASET
bcadmin -c . darc add --out_id data_owner_1/movies/darc_id.txt --out_key data_owner_1/movies/darc_key.txt --desc "DARC for the movies dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_1/movies/darc_id.txt) --id $(cat data_owner_1/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_1/movies/darc_id.txt) --id $(cat data_owner_1/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_1/movies/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_1/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 1 should be able to update the darc for the movie dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_1/movies/darc_id.txt) --id "$(cat data_owner_1/darc_id.txt) | attr:allowed:uses=clientData,financeData&purposes=regulatory,singleCounterParty,multipleCounterparty,essentialBusinessFunction" --sign $(cat data_owner_1/darc_key.txt) --replace --restricted

# Add the movie dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Movies actors",
	"Description": "Images of celebrities from various movies",
	"CloudURL": "dedis/datasets/2_images.tar.aes",
	"Author": "Michel Fritz",
	"WriteInstID": "",
	"SHA2": "bdf7daecaf76d97ef0fed0fe3230a852ddb1814cf0ad20680f382ec29b6ca5e5"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_1/movies/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_1/darc_key.txt) --instid $(cat lts_id.txt) --secret 4d0479cb26816af63f9fc997c3bcbf7e50baf4992e18e925c6e4eb72 --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_1/movies/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_1/movies/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)



# Darc for a SECOND OWNER

bcadmin -c . darc add --out_id data_owner_2/darc_id.txt --out_key data_owner_2/darc_key.txt --desc "DARC of the data owner N°2"

# Darc for the CGR Dataset
bcadmin -c . darc add --out_id data_owner_2/cgr/darc_id.txt --out_key data_owner_2/cgr/darc_key.txt --desc "DARC for the cgr dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_2/cgr/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/cgr/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_2/cgr/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_2/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 2 should be able to update the darc for the cgr dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/cgr/darc_id.txt) --id "$(cat data_owner_2/darc_id.txt) | attr:allowed:uses=clientData&purposes=regulatory,singleCounterParty" --sign $(cat data_owner_2/darc_key.txt) --replace --restricted

# Add the cgr dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Claim Growth Rate IN Basel (CGR)",
	"Description": "Claim growth rate in different quarters of Canton Basel-Stadt. It is correlated with the egr_data. This is a private dataset that can be used only for fiancial purpose and single counterparty use.",
	"CloudURL": "dedis/datasets/cgr_data.csv.aes",
	"Author": "Christian Lazzari",
	"WriteInstID": "",
	"SHA2": "c0e050868b99c600db09963205768971958d3793862e0b7acf5fb9d0cf43a630"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_2/cgr/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_2/darc_key.txt) --instid $(cat lts_id.txt) --secret d99fc6af75205dc29670f7fdbec8dcff0432a5965185c4271bcf6647 --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_2/cgr/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_2/cgr/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)



# Darc for the EGR Dataset
bcadmin -c . darc add --out_id data_owner_2/egr/darc_id.txt --out_key data_owner_2/egr/darc_key.txt --desc "DARC for the egr dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_2/egr/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/egr/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_2/egr/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_2/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 2 should be able to update the darc for the egr dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/egr/darc_id.txt) --id "$(cat data_owner_2/darc_id.txt) | attr:allowed:uses=clientData&purposes=regulatory,singleCounterParty" --sign $(cat data_owner_2/darc_key.txt) --replace --restricted

# Add the egr dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Employment Growth Rate in Basel (EGR)",
	"Description": "Employment growth rate in different quarters of Canton Basel-Stadt. This is a public dataset that has no restrictions.",
	"CloudURL": "dedis/datasets/egr_data.csv.aes",
	"Author": "Christian Lazzari",
	"WriteInstID": "",
	"SHA2": "ede087172dbfb94898b372159f376235df8f8705be94fc1a369d7ab5f4d00857"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_2/egr/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_2/darc_key.txt) --instid $(cat lts_id.txt) --secret d5616d9349510837d7803321cf803a273783f010955e3614ce900b7d --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_2/egr/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_2/egr/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)



# Darc for the WE_Wohnviertel Dataset
bcadmin -c . darc add --out_id data_owner_2/WE_Wohnviertel/darc_id.txt --out_key data_owner_2/WE_Wohnviertel/darc_key.txt --desc "DARC for the WE_Wohnviertel dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_2/WE_Wohnviertel/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/WE_Wohnviertel/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_2/WE_Wohnviertel/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_2/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 2 should be able to update the darc for the WE_Wohnviertel dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/WE_Wohnviertel/darc_id.txt) --id "$(cat data_owner_2/darc_id.txt) | attr:allowed:uses=clientData&purposes=regulatory,singleCounterParty" --sign $(cat data_owner_2/darc_key.txt) --replace --restricted

# Add the WE_Wohnviertel dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Basel boundaries",
	"Description": "Boundaries of quarters of Canton Basel-Stadt in GeoJSON format. Public dataset that has no restrictions.",
	"CloudURL": "dedis/datasets/WE_Wohnviertel.json.aes",
	"Author": "Christian Lazzari",
	"WriteInstID": "",
	"SHA2": "9d3e83c88702a1562227f4cc74676776c1b543863773db179fdcefa3fa2109d7"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_2/WE_Wohnviertel/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_2/darc_key.txt) --instid $(cat lts_id.txt) --secret 53f80ceb6c968f77b2d639cd5fdb56ee6c83d85369e3cbbb71346bf3 --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_2/WE_Wohnviertel/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_2/WE_Wohnviertel/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)



# Darc for the BCV (Basel Claim Value) Dataset
bcadmin -c . darc add --out_id data_owner_2/bcv/darc_id.txt --out_key data_owner_2/bcv/darc_key.txt --desc "DARC for the bcv dataset" -id $(cat darc_key.txt)  --unrestricted
bcadmin -c . darc rule --rule "spawn:calypsoWrite" --darc $(cat data_owner_2/bcv/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/bcv/darc_id.txt) --id $(cat data_owner_2/darc_id.txt) --sign $(cat darc_key.txt)
bcadmin -c . darc rule --rule "invoke:darc.evolve" --darc $(cat data_owner_2/bcv/darc_id.txt) --id "$(cat darc_key.txt) | $(cat data_owner_2/darc_id.txt)" --replace --sign $(cat darc_key.txt)

# Now the data owner 2 should be able to update the darc for the bcv dataset (in restricted mode only)
bcadmin -c . darc rule --rule "spawn:calypsoRead" --darc $(cat data_owner_2/bcv/darc_id.txt) --id "$(cat data_owner_2/darc_id.txt) | attr:allowed:uses=&purposes=regulatory" --sign $(cat data_owner_2/darc_key.txt) --replace --restricted

# Add the bcv dataset on the catalog

read -r -d '' DATA << EOM
{
	"Title": "Basel Claim Values",
	"Description": "Claim values (in CHF) for different regions in Basel-Stadt. Agreement to use this dataset is not yet established and we cannot use it.",
	"CloudURL": "dedis/datasets/bcv_data.csv.aes",
	"Author": "Christian Lazzari",
	"WriteInstID": "",
	"SHA2": "529cec07ea605e863ae58c95b51148ac0f156c27d26f7b9c5fd600ec27aa9db3"
}
EOM

WRITE_ID=$(csadmin -c . contract write spawn -x --darc $(cat data_owner_2/bcv/darc_id.txt) --extraData "$DATA" --sign $(cat data_owner_2/darc_key.txt) --instid $(cat lts_id.txt) --secret d9b6e688a5d9f25e8ff7569a2a222707040aa698310d18c7c24bf0d5 --key $(cat pub_key.txt))
echo $WRITE_ID > data_owner_2/bcv/write_id.txt

CONTENT=$(bcadmin -c . contract value get -i $(cat catalog_id.txt))
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$CONTENT,$(cat data_owner_2/bcv/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)

# update the catalog at once
bcadmin -c . contract value invoke update -i $(cat catalog_id.txt) --value "$(cat data_owner_2/bcv/write_id.txt),$(cat data_owner_2/cgr/write_id.txt),$(cat data_owner_2/egr/write_id.txt),$(cat data_owner_2/WE_Wohnviertel/write_id.txt),$(cat data_owner_1/titanic/write_id.txt),$(cat data_owner_1/movies/write_id.txt)" --darc $(cat darc_id.txt) --sign $(cat darc_key.txt)


#
# DARC for the data scientist
#
mkdir data_scientist_1
bcadmin -c . darc add --out_id data_scientist_1/darc_id.txt --out_key data_scientist_1/darc_key.txt --desc "DARC for the data scientist 1"  --unrestricted

bcadmin darc rule -rule "spawn:odysseyproject" -darc $(cat data_scientist_1/darc_id.txt) -sign $(cat data_scientist_1/darc_key.txt) -identity $(cat data_scientist_1/darc_key.txt)
bcadmin darc rule -rule "invoke:odysseyproject.updateMetadata" -darc $(cat data_scientist_1/darc_id.txt) -sign $(cat data_scientist_1/darc_key.txt) -identity $(cat data_scientist_1/darc_key.txt)
# We need both the enclave manager and the data scientist to update the status
# because the data scientist sets the status when it updates the attributes of
# the project and the enclave manager sets all the other statuses (preparing,
# unlocking, destroying).
bcadmin darc rule -rule "invoke:odysseyproject.updateStatus" -darc $(cat data_scientist_1/darc_id.txt) -sign $(cat data_scientist_1/darc_key.txt) -identity "$(cat darc_key.txt) | $(cat data_scientist_1/darc_key.txt)" --replace
bcadmin darc rule -rule "invoke:odysseyproject.setURL" -darc $(cat data_scientist_1/darc_id.txt) -sign $(cat data_scientist_1/darc_key.txt) -identity $(cat darc_key.txt)
# the access pub key is set during the spawn
# bcadmin darc rule -rule "invoke:odysseyproject.setAccessPubKey" -darc $ID -sign $KEY -identity $KEY
bcadmin darc rule -rule "invoke:odysseyproject.setEnclavePubKey" -darc $(cat data_scientist_1/darc_id.txt) -sign $(cat data_scientist_1/darc_key.txt) -identity $(cat darc_key.txt)



#
# Create the catalog
#

export BC="/Users/nkocher/GitHub/odyssey/secret/bc-6fd9a73d377e3e9150babbaabc83a2f7a7bd57f2a8911d454f02d9a43b206295.cfg"
export BC_CONFIG="/Users/nkocher/GitHub/odyssey/secret/"

bcadmin darc rule --sign $(cat darc_key.txt) --darc $(cat darc_id.txt) -id $(cat darc_id.txt) -rule "spawn:odysseycatalog"
bcadmin darc rule --sign $(cat darc_key.txt) --darc $(cat darc_id.txt) -id $(cat darc_id.txt) -rule "invoke:odysseycatalog.addOwner"
bcadmin darc rule --sign $(cat darc_key.txt) --darc $(cat darc_id.txt) -id $(cat darc_id.txt) -rule "invoke:odysseycatalog.odysseycatalog.updateMetadata"
bcadmin darc rule --sign $(cat darc_key.txt) --darc $(cat darc_id.txt) -id $(cat darc_id.txt) -rule "invoke:odysseycatalog.deleteDataset"
bcadmin darc rule --sign $(cat darc_key.txt) --darc $(cat darc_id.txt) -id $(cat darc_id.txt) -rule "invoke:odysseycatalog.archiveDataset"
catadmin contract catalog spawn --darc $(cat darc_id.txt) --sign $(cat darc_key.txt) 
catadmin contract catalog invoke addOwner --firstname Noémien --lastname Kocher -i 4f541c8a70f2832ea7b96af31f09604612e3e5465614175439eaa7516c968e31 --sign $(cat darc_key.txt) --identityStr $(cat data_owner_1/darc_key.txt)

read -r -d '' JSONATTR << EOM
{
	"attributesGroups": [{
			"title": "Use",
			"description": "How this dataset can be used",
			"consumer_description": "Please tell us how the result will be used",
			"attributes": [{
					"id": "use_restricted",
					"name": "use_restricted",
					"description": "The use of this dataset is restricted",
					"type": "checkbox",
					"rule_type": "must_have",
					"delegated_enforcement": true,
					"attributes": [{
						"id": "use_restricted_description",
						"name": "use_restricted_description",
						"description": "Please describe the restriction",
						"type": "text",
						"rule_type": "must_have",
						"delegated_enforcement": true,
						"attributes": []
					}]
				},
				{
					"id": "use_retention_policy",
					"name": "use_retention_policy",
					"description": "There is a special retention policy for this dataset",
					"type": "checkbox",
					"rule_type": "must_have",
					"delegated_enforcement": true, 
					"attributes": [{
						"id": "use_retention_policy_description",
						"name": "use_retention_policy_description",
						"description": "Please describe the retention policy",
						"type": "text",
						"rule_type": "must_have",
						"delegated_enforcement": true,
						"attributes": []
					}]
				},
				{
					"id": "use_predefined_purpose",
					"name": "use_predefined_purpose",
					"description": "Can be used for the following predefined-purposess",
					"consumer_description": "the result will be used for the following pre-defined purposes",
					"type": "checkbox",
					"rule_type": "allowed",
					"attributes": [{
						"id": "use_predefined_purpose_legal",
						"name": "use_predefined_purpose_legal",
						"description": "To meet SR's legal or regulatory requirements",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_counterparty",
						"name": "use_predefined_purpose_analytics_counterparty",
						"description": "To provide analytics to the counterparty of the contract from which data is sourced/for the counterparty's benefit",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_essential",
						"name": "use_predefined_purpose_essential",
						"description": "To perform an essential function of SR's business (for example, reserving, develop and improve costing/pricing models, portfolio management, enable risk modelling and accumulation control)",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_analytics_multiple_counterparty",
						"name": "use_predefined_purpose_analytics_multiple_counterparty",
						"description": "To provide analytics to multiple counterparties / for the benefit of multiple counterparties",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_general_information",
						"name": "use_predefined_purpose_general_information",
						"description": "To provide general information to the public",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}, {
						"id": "use_predefined_purpose_sole_sr",
						"name": "use_predefined_purpose_sole_sr",
						"description": "For the sole benefit of SR",
						"type": "checkbox",
						"rule_type": "allowed",
						"attributes": []
					}]
				}
			]
		},
		{
			"title": "Classification",
			"description": "The following classification types apply on my dataset",
			"consumer_description": "I have the right to work with the following types of data",
			"attributes": [{
				"id": "classification_public",
				"name": "classification_public",
				"description": "Public information",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_internal",
				"name": "classification_internal",
				"description": "Internal data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_confidential",
				"name": "classification_confidential",
				"description": "Confidential data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_critical",
				"name": "classification_critical",
				"description": "Critical data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "classification_personal",
				"name": "classification_personal",
				"description": "Personal data",
				"type": "checkbox",
				"rule_type": "must_have",
				"attributes": []
			}]
		},
		{
			"title": "Access",
			"description": "Tell us who can access the data",
			"consumer_description": "Please select who will have access the the result",
			"attributes": [{
				"id": "access_unrestricted",
				"name": "access",
				"description": "No restriction",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "access_internal",
				"name": "access",
				"description": "SR internal",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": []
			}, {
				"id": "access_defined_group",
				"name": "access",
				"description": "SR defined group",
				"type": "radio",
				"rule_type": "must_have",
				"attributes": [{
					"id": "access_defined_group_description",
					"name": "access_defined_group_description",
					"description": "Please specify the group",
					"delegated_enforcement": true,
					"type": "text",
					"rule_type": "must_have"
				}]
			}]
		}
	],
	"delegated_enforcement": {
		"title": "Manual enforcement",
		"description": "some restriction can not be automatically checked. Therefore, you are requested to agree on the following attributes.",
		"attributes": [{
			"id": "use_restricted_description_enforcement",
			"description": "I agree on this restriction use",
			"value_from_id": "use_restricted_description",
			"trigger_id": "use_restricted",
			"trigger_value": "",
			"check_validates": [
				"use_restricted"
			],
			"text_validates": "use_restricted_description"
		},{
			"id": "use_retention_policy_description_enforcement",
			"description": "I agree on this retention policy",
			"value_from_id": "use_retention_policy_description",
			"trigger_id": "use_retention_policy",
			"trigger_value": "",
			"check_validates": [
				"use_retention_policy"
			],
			"text_validates": "use_retention_policy_description"
		},{
			"id": "access_defined_group_description_enforcement",
			"description": "I certify that the result will only be used by this specific group from SR",
			"value_from_id": "access_defined_group_description",
			"trigger_id": "access_defined_group",
			"trigger_value": "",
			"text_validates": "access_defined_group_description"
		}]
	}
}
EOM

catadmin contract catalog invoke updateMetadata -i 4f541c8a70f2832ea7b96af31f09604612e3e5465614175439eaa7516c968e31 --metadataJSON "$JSONATTR" --sign $(cat darc_key.txt)

# create the config file

cd darc_owner_1
bcadmin -c . link ../roster.toml 6fd9a73d377e3e9150babbaabc83a2f7a7bd57f2a8911d454f02d9a43b206295 --darc $(cat darc_id.txt) --id $(cat darc_key.txt)