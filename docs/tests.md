# Tests

The system is only at an early maturity stage in term of testing. Some efforts
have been invested in preparing the ground for testing. Therefore, the DOManager
has a first batch of tests that demonstrates how to mock the http server, the
commands performed with os.exec, the external cloud provider and the task
manager. It is a good start to develop more tests across the system. Some other
elements like the catalog smart contract have been well tested. This shows the
way for testing the project smart contract.

A first batch of tests have been added in the DSManager. It uses the same
mocking interfaces as the DOManager and adds a new one to mock the calls to the
Enclave Manager.

Rating:  
0 = Not tested  
1 = Barely tested  
2 = Mostly tested  
3 = Fully tested  

| Module | Testing level | Remarks |
|--------|---------------|--------|
| Catalogc | 🌕🌕🌗 2.5  | catadmin not tested |
| Cryptutil | 🌕🌕🌕 3 | |
| DOManager | 🌕🌗🌑 1.5 | can be used as a base for the test in DSManager |
| DSManager | 🌕 1 | |
| Enclave | 0 | |
| Enclavem | 0 | |
| Projectc | 0 | |

You can launch all the tests with:

```make
make test
```

## Components

Decoupling our system with clear interfaces and components gave us a nice ground
for testing. Most of the interfaces needed to mock the different components are
available in `dsmanager/app/helpers`.

The following illustration shows the different interfaces required to mock the
components:

![](assets/components_uml_tests.png)