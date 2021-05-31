---
name: Add New Chain Template
about: Checklist of things that need to be done before a chain can be enabled
title: "[Checklist] Enable <Chain Name>"
labels: ''
assignees: ''

---

# [Checklist] Enable <Chain Name>

- [ ] Fund relayer account with funds on the new chain 
- [ ] Add chain entry to CNS
- [ ] Add chain to the operator (full-node, relayer, ...)
- [ ] Check that chain_status endpoint works and that the full-node is synced 
- [ ] Manually verify denoms that need to be verified
- [ ] Make sure at least one verified denom has a `fee` field set up with all three value  
- [ ] Make sure the primary channels are correctly set up 
- [ ] Make sure price API is set up with newly added verified denoms 
- [ ] Set relayer account alerting limit 

### Only check this if all the box above have been checked

- [ ] Mark the chain as "enabled" in the CNS
