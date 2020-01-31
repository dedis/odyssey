# Auditable Sharing and Management of Sensitive Data Across Jurisdictions

Because sharing sensitive data between multiple distrustfull parties can
be a challenge, Odyssey makes use of state-of-the-art secret management
on the blockchain, coupled with an enclave deleviery mechanism, to
ensure controlled and safe delivery of the data, as well as proper
destruction with controller life-cycle. At rest, data are stored
encrypted on a private cloud provider. Data can be requested and
decrypted based on the attribute of a project that clearly defines the
context on wich the data will be used. Data are never decrypted outside
a virtual machine created on fly for that purpose. The lifecycle of the
VM ensures that unencrypted data are deleted after use, preventing
accidental or malicious leakage.

This repo holds all the components necessary to run the Odyssey
projects. You will find 3 components:

- **Data Scientist Manager**, user application that delivers requested datasets to an encalve
- **Data Owner Manager**, user application that allows one to upload and update datasets
- **Enclave Manager**, server application that handles the lifecycle of enclaves

Additionally, some tools were needed to support the system:

- **Projectc**, a smart contract holding the attributes of a project
- **Catalogc**, a smart contract holding the catalog of available datasets and defines the attributes to control them
- **Cryptutil**, a command line tool to encrypt and decrypt data with AES-CGM
- **Enclave**, scripts used on the enclave
