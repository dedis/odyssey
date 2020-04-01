# Auditable Sharing and Management of Sensitive Data Across Jurisdictions

<center>
<img src="assets/odyssey-components.png">
</center>

Odyssey is a set of applications and tools that enables the sharing sensitive
data between multiple distrustfull parties. This project uses state-of-the-art
secret management service [1] on the blockchain [2] coupled with an enclave
delivery mechanism. This uniq combination provides auditable access on shared
datasets, collective agreement, and controled life-cycle of the data that
prevents malicious or accidental leakage of data. At rest, data is stored
encrypted on a private cloud provider. Data can be requested and decrypted based
on the attributes of a project that clearly defines the context on wich the data
will be used. Those attributes are stored on the blockchain and the data is
released when a quorum of nodes agree that the attributes of the project comply
with the defined use of the data. The data is never decrypted outside a virtual
machine (VM) created on fly for that purpose. The lifecycle of the VM ensures
that unencrypted data is deleted after use, preventing accidental or malicious
leakage.

[1] [Calypso - Auditable Sharing of Private Data over Blockchains](https://eprint.iacr.org/2018/209)  
[2] [OmniLedger: A Secure, Scale-Out, Decentralized Ledger via Sharding](https://eprint.iacr.org/2017/406)

This repo holds all the components necessary to run the Odyssey projects. You
will find 3 components:

- **Data Scientist Manager**, user application that delivers requested datasets
  to an encalve
- **Data Owner Manager**, user application that allows one to upload and update
  datasets
- **Enclave Manager**, server application that handles the lifecycle of enclaves

![DSM logo](assets/dsm-logo.png) ![DOM logo](assets/dom-logo.png) ![ENM
logo](assets/enm-logo.png)

Additionally, some tools were needed to support the system:

- **Projectc**, a smart contract holding the attributes of a project
- **Catalogc**, a smart contract holding the catalog of available datasets along
  with their attributes that control their acess
- **Cryptutil**, a command line tool to encrypt and decrypt data with AES-CGM
- **Enclave**, scripts used on the enclave (ie. VMs)



## Demo

Here is a 15 minutes video that walks throught the whole process of uploading a
dataset to its use via a secure enclave.

[see the
video](https://drive.google.com/file/d/1QBvqjBjUS3q0Z9CShm4pR7lBC6wbj6cw/view)

## Screenshots

### Data Owner Manager (upload and management of datasets)

<figure>
    <figcaption>Welcome page</figcaption>
    <img src="assets/screenshots/dom7.png">
</figure>

<figure>
    <figcaption>Upload of a dataset</figcaption>
    <img src="assets/screenshots/dom1.png">
</figure>

<figure>
    <figcaption>Task created to upload a dataset</figcaption>
    <img src="assets/screenshots/dom3.png">
</figure>

<figure>
    <figcaption>List of datasets</figcaption>
    <img src="assets/screenshots/dom2.png">
</figure>

<figure>
    <figcaption>Edition of dataset: General infos</figcaption>
    <img src="assets/screenshots/dom4.png">
</figure>

<figure>
    <figcaption>Edition of dataset: Attributes</figcaption>
    <img src="assets/screenshots/dom5.png">
</figure>

<figure>
    <figcaption>Edition of dataset: Special actions and DARC</figcaption>
    <img src="assets/screenshots/dom6.png">
</figure>

<figure>
    <figcaption>Audit of a dataset access</figcaption>
    <img src="assets/screenshots/dom8.png">
</figure>

<figure>
    <figcaption>Lifecycle of a project</figcaption>
    <img src="assets/screenshots/dom9.png">
</figure>

<figure>
    <figcaption>Lifecycle of a project (enclave destruction)</figcaption>
    <img src="assets/screenshots/dom10.png">
</figure>

### Data Scientist Manager (use of datasets on enclaves)

<figure>
    <figcaption>Welcome page</figcaption>
    <img src="assets/screenshots/dsm1.png">
</figure>

<figure>
    <figcaption>Project creation: selection of a dataset</figcaption>
    <img src="assets/screenshots/dsm7.png">
</figure>

<figure>
    <figcaption>Project page</figcaption>
    <img src="assets/screenshots/dsm2.png">
</figure>

<figure>
    <figcaption>Request to create the project: enclave initialization</figcaption>
    <img src="assets/screenshots/dsm3.png">
</figure>

<figure>
    <figcaption>A failed attempty to unlock the enclave</figcaption>
    <img src="assets/screenshots/dsm6.png">
</figure>

<figure>
    <figcaption>Attribute update on a project after a failed attempt</figcaption>
    <img src="assets/screenshots/dsm5.png">
</figure>

<figure>
    <figcaption>Access page of the enclave (after successful unlock)</figcaption>
    <img src="assets/screenshots/dsm4.png">
</figure>

<style>
  figcaption {
    text-align: center;
  }
</style>