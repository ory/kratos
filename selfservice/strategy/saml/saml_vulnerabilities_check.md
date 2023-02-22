# **SAML Crewjam over Kratos vulnerabilities checks**

|Status|Summary|Description|Comment|
| :- | :- | :- | :- |
|**OK**|**Check that it's not possible to modify the signed SAML Response**|||
|OK|- By adding an attribute|||
|OK|- By adding an element|||
|OK|- By modifying the indent|||
|**OK**|**Check that it's not possible to modify the signed SAML Assertion**|**If the SAML Response isn't signed**||
|OK|- By adding an attribute|||
|OK|- By adding an element|||
|**OK**|**Check that it's not possible to remove the signature**|||
|OK|- Not possible to remove SAML Response signature value|A signature must contain a signature value||
|OK|- Possible to remove SAML Response signature if the SAML Assertion is signed|Either the SAML Response or the SAML Assertion must be signed||
|OK|- Not possible to remove SAML Assertion signature value|A signature must contain a signature value||
|OK|- Not possible to remove SAML Assertion signature|If the SAML Response is still signed, any SAML Assertion modification is an unauthorized SAML Response modification||
|OK|- Not possible to remove both SAML Response signature and SAML Assertion signature|Either the SAML Response or the SAML Assertion must be signed||
|**OK**|**Prevent from Signature Wrapping Attacks (XSW)**|||
|OK|- XSW1 response wrap 1|XSW #1 manipulates SAML Responses. It does this by making a copy of the SAML Response and Assertion, then inserting the original Signature into the XML as a child element of the copied Response. The assumption being that the XML parser finds and uses the copied Response at the top of the document after signature validation instead of the original signed Response.||
|OK|- XSW2 response wrap 2|Similar to XSW #1, XSW #2 manipulates SAML Responses. XSW #1 and XSW #2 are the only two that deal with Responses. The key difference between #1 and #2 is that the type of Signature used is a detached signature where XSW #1 used an enveloping signature. The location of the malicious Response remains the same.||
|OK|- XSW3 assertion wrap 1|XSW #3 is the first example of an XSW that wraps the Assertion element. SAML Raider inserts the copied Assertion as the first child of the root Response element. The original Assertion is a sibling of the copied Assertion.||
|OK|- XSW4 assertion wrap 2|XSW #4 is similar to #3, except in this case the original Assertion becomes a child of the copied Assertion.||
|OK|- XSW5 assertion wrap 3|XSW #5 is the first instance of Assertion wrapping we see where the Signature and the original Assertion aren’t in one of the three standard configurations (enveloped/enveloping/detached). In this case, the copied Assertion envelopes the Signature.||
|OK|- XSW6 assertion wrap 4|XSW #6 inserts its copied Assertion into the same location as #’s 4 and 5. The interesting piece here is that the copied Assertion envelopes the Signature, which in turn envelopes the original Assertion.||
|OK|- XSW7 assertion wrap 5|XSW #7 inserts an Extensions element and adds the copied Assertion as a child. Extensions is a valid XML element with a less restrictive schema definition. OpenSAML used schema validation to correctly compare the ID used during signature validation to the ID of the processed Assertion. The authors found in cases where copied Assertions with the same ID of the original Assertion were children of an element with a less restrictive schema definition, they were able to bypass this particular countermeasure.||
|OK|- XSW8 assertion wrap 6|XSW #8 uses another less restrictive XML element to perform a variation of the attack pattern used in XSW #7. This time around the original Assertion is the child of the less restrictive element instead of the copied Assertion.||
|**OK**|**Analyse the application behaviour when adding XML comments**|**In the beginning, middle, and end of an attribute (such as username)**||
|OK|- The XML comments aren't removed|||
|OK|- The XML comments don't allow the user to authenticate with another identity|||
|**OK**|**Prevent from signing the SAML Response with own certificate**|**Depending on the case: assertion, message, or both**||
|OK|- Prevent from signing the SAML assertion with own certificate|||
|OK|- Prevent from signing the SAML response with own certificate|||
|OK|- Prevent from signing both the response and the assertion|||
|**OK**|**Prevent from XXE and XSLT attacks**|||
|**OK**|**Check if there are any known vulnerabilities for the SAML library or software in use**|||
|**OK**|**Check if it is possible to send the same SAML Response twice (Replay Attack)**|||
|**OK**|**Check if the SP uses the same attribute as IdP to identify the user**||**There is a mapping**|
|**N/A**|**Check if IdP allows anonymous registration**|||
|**N/A**|**Verify Single Log Out (if required)**|||
|**OK**|**Check if the validity time window is short**|**3-5 minutes**|**90sec in our implementation**|
|**OK**|**Check if the time window is validated**|**Try to use the same SAML Response after it has expired**||
|**N/A**|**Check for Cross-Site Request Forgery attack**|**Unsolicited Response**||
|**OK**|**Check if the recipient is validated**|**Token Recipient Confusion**||
|**N/A**|**Check for Open Redirect in RelayState**|||
|**OK**|**Check the signature algorithm in use**||**SHA256**|
|**OK**|**Check that the SAML response is associated with an AuthnRequest already performed on the IdP**|**Check that the ID field of the SAML Request corresponds to the InResponseTo field of the SAML Response**||

## **Sources**
- [SAML – what can go wrong? Security check](https://www.securing.pl/en/saml-what-can-go-wrong-security-check/)
- [How to Hunt Bugs in SAML; a Methodology - Part II](https://epi052.gitlab.io/notes-to-self/blog/2019-03-13-how-to-test-saml-a-methodology-part-two/)
- [On Breaking SAML: Be Whoever You Want to Be](https://www.usenix.org/system/files/conference/usenixsecurity12/sec12-final91.pdf)
- [SAMLRaider](https://github.com/CompassSecurity/SAMLRaider)
