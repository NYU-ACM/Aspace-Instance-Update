# Aspace-Instance-Update
This application is used to batch update Top Container URIs (box numbers) and Child Indicators (folder numbers) for Archival Object's within a Resource. The application uses the work order output from the [Hudson Molonglo's Digitization Work Order Plugin](https://github.com/hudmol/digitization_work_order).

usage<br/>
`$ aspace-instance-update --workorder your-workorder.tsv --environment your-environment`

Options
-------
* --workorder, required,	/path/to/workorder.tsv
* --environment, required,      aspace environment to be used: dev/stage/prod
* --undo, optional,	runs a work order in revrse, undo a previous run
* --test, optional,	test mode does not execute any POSTs, this is recommended before running on any data
* --help	print this help message

Work-Order Specification
------------------------
| Resource ID	| Ref ID	| URI	| Container Indicator 1	| Container Indicator 2	| Container Indicator 3	| Title	| Component ID |
| ---	| ---	| ---| ---	| --- | --- | ---	| --- |
| TAM.011	| ref14	| /repositories/2/archival_objects/154967	| 1 | 	1 |  | Correspondence	| |

In a spreadsheet editor add two columns to the work order: 'New Container Indicator 1' and 'New Container Indicator 2'. The updater will update the Instances updating the reference to the top container and the child indicator.

| Resource ID	| Ref ID	| URI	| Container Indicator 1	| Container Indicator 2	| Container Indicator 3	| Title	| Component ID | New Container Indicator 1	| New Container Indicator 2 |
| ---	| ---	| ---| ---	| --- | --- | ---	| --- | ---	| --- |
| TAM.011	| ref14	| /repositories/2/archival_objects/154967	| 1 | 	1 |  | Correspondence	| | 2 | 1 |

