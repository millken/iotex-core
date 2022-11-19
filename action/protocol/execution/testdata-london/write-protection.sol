// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

// NOTE: Deploy this contract first
contract B {
    // NOTE: storage layout must be the same as contract A
    uint public num;
    address public sender;
    uint public value;
    event Done();

    function setVars(uint _num) public payable {
        num = _num;
        sender = msg.sender;
        value = msg.value;
        emit Done();
    }
}

contract A {
    uint public num;
    address public sender;
    uint public value;
    address private c;
    mapping(uint => uint) private _a;
    event Success(bool);
    event LOG(uint);

    function make() public {
        c = address(new B());
        _a[1] = 1;
        _a[2] = 2;
    }

    function callStatic(address _contract,uint _num) public {
        if (_num == 0) {
            _contract.call(
                abi.encodeWithSignature("notfund()")
            );
            LOG(0);
            _contract.call(
                abi.encodeWithSignature("notfund2()")
            );
            LOG(1);
            _contract.call(
                abi.encodeWithSignature("notfund3()")
            );
            LOG(2);
        }
        if (_num < 2) {
            delete _a[1];
            delete _a[2];
            (bool success, bytes memory _) = _contract.staticcall(
                abi.encodeWithSignature("setVars(uint256)", _num)
            );
         emit Success(success);
       }
        if (_num < 2) {
            revert();
        }
    }

    function setVars(address _contract, uint _num) public returns (uint256) {
        if (_contract == address(0)) {
            _contract = c;
        }
        try this.callStatic(_contract,0) {
        } catch {
            try this.callStatic(_contract,1) {
            }catch{
                this.callStatic(_contract,2);
            }
        }

        return 1;
    }
}