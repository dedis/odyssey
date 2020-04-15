# Design

Note: one may first need to be familiar with the terminology used in
cothority. Please refer to the documentation in the
[cothority repos](https://github.com/dedis/cothority)
to learn more.

## Threat model

### Target environment

We assume that there is a central IT governance model, and support for
security and auditing of network activity. The IT department knows who
the authorised users of the network are, and takes responsibility for
authenticating them via a Single Sign-On system, and authorising access
to apps via some kind of access control list.  The IT department is also
capable of and responsible for operating a hypervisor which will be used
to spawn, run, and destroy enclaves. In our system, this hypervisor is
the enclave manager.

### Authorised users

Authorised users are expected to not be malicious. They are motivated by
the commitment to their job, and share the goals of the organisation,
and they are aware that malicious behavior will be investigated and
result in consequences including loss of job and possibly criminal
referral to the police. But, they are human, and make human errors. The
security offered by the system should serve them by protecting them from
errors that could leak data, and should serve the audit function of the
organisation by collecting evidence of wrongdoing.

# Figures

**Logical view, cothority centric**
<center><img src="assets/cothority_view.png"/></center>

**UML component diagram**
<center><img src="assets/components_uml.png"/></center>

# About the code structure

## Global organization

Each component has its own folder at the root of the project. 

- `catalogc/` the catalog smart contract
- `cryptutil/` utility tool to encrypt/descript data
- `domanager/` the data owner manager
- `dsmanager/` the data scientist manager
- `enclave/` scipts that run on each new enclave
- `enclavem/` the enclave manager
- `ledger/` the code taken from [cothority](https://github.com/dedis/cothority) to run a blockchain
- `projectc/` the project smart contract

## Http servers

`domanager/`, `dsmanager/`, and `enclavem/` are all http servers that use the
native net/http library, coupled with some
[gorilla](http://www.gorillatoolkit.org) packages. All our http components
follow a standard organization borrowed from the [ruby on
rails](https://guides.rubyonrails.org/getting_started.html) framework:

- `main.go` the entrypoint of the server that defines the routes and their corresponding handlers
- `assets/` images, fonts, stylesheets and all other asset elements needed
- `controllers/` contains the handlers that are mapped from the routes
- `helpers/` common utilities needed mostly in the controllers
- `models/` codes that support the different data model structures needed
- `views/` html representations that are rendered from the controllers

It came out that some elements of the helpers were common to all our http
servers. Instead of duplicating the codes accross the different helpers, we
decided that the helpers of the data scientist manager (`dsmanager/app/helpers`)
would contain the common helpers. This is for example the case of the "Task"
helper, which is used by all the 3 http servers.

# About the DARCs

In order to later display the content of each DARC, the following table sets the
vocabulary of each identites and DARCs:

| Entity | DARC | Identity |
| ------ | ---- | -------- |
| 🔬 Data scientist | `darc(🔬)` | `id(🔬)` |
| 🐙 Enclave manager | `darc(🐙)` | `id(🐙)` |
| 👔 Data owner | `darc(👔)` | `id(👔)` |
| 📦 Dataset | `darc(📦)` | - |
| 🔐 Enclave | - | `id(🔐)` |

## darc(🔬) - Data scientist

Rationale: The data scientist is responsible for creating the project instance
and setting the attributes on it. However, only the enclave manager has the
control over the DARC and can set the URL and public key of the enclave.

| Action | Rule | 
| ------ | ---- |
| `invoke:darc.evolve` | `id(🐙)` |
| `spawn:odysseyproject` | `id(🔬)` |
| `invoke:odysseyproject.update` | `id(🔬)` |
| `invoke:odysseyproject.updateStatus` | `id(🔬) \| id(🐙)` |
| `invoke:odysseyproject.setURL` | `id(🐙)` |
| `invoke:odysseyproject.setEnclavePubKey` | `id(🐙)` | 

## darc(🐙) - Enclave manager

Rationale: The enclave manager doesn't need a DARC with a lot of rules on it
because it creates itself the other DARCs and can then ensure it has the correct
rights on them. For example, it creates the DARC for each dataset and ensure it
can create a read request.

| Action | Rule | 
| ------ | ---- |
| `spawn:darc` | `id(🐙)` |
| `invoke:darc.evolve` | `id(🐙)` |
| `spawn:odysseycatalog` | `id(🐙)` |
| `invoke:odysseycatalog.addOwner` | `id(🐙)` |

## darc(👔) - Data owner

Rationale: The data owner's DARC mainly servers as an identity proxy. It is
convenient because it allows a data owner to update its identity and still get
access to its datasets. To do this it only needs to update the `_sign` action
(Note: the `_sign` actions are omited because they always have the default rule
on them, which is the identity of the entity that owns the DARC).

| Action | Rule | 
| ------ | ---- |
| `invoke:darc.evolve` | `id(👔)` |

## darc(📦) - Dataset

Rationale: The data owner is able to create a write and read request, while the
enclave manager can create a read request provided it has the right project's
attributes.

| Action | Rule | 
| ------ | ---- |
| `invoke:darc.evolve` | `darc(👔) \| id(🐙)` |
| `spawn:calypsoWrite` | `darc(👔)` |
| `spawn:calypsoRead` | `darc(👔) \| (id(🐙) & attr:<custom_attributes>)` |

## id(🔐) - Enclave

The enclave's ID is not directly used in DARCs because it is only needed by the
enclave manager to create the read request.
