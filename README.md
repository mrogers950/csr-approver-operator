# csr-approver-operator

WIP : This operator manages a CSR API approval controller.
 - Watch the CSR endpoint for CSR requests
 - Decide if the CSR should be allowed or denied
   - Based on requestor permissions / allowed usages / allowed names / allowed hostnames
 - Approve or deny and update CSR status
