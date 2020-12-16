# Amazon S3 Zipping tool (s3packer)

## What does it do?

### 1. Zips S3 files

Takes an amazon s3 bucket folder and zips it to a:

- Local File
- S3 File (ie uploads the zip back to s3)

### Installation

### Usage

Zip contents into an S3 bucket, does not have to be the origin of the S3 data

    s3packer s3://bucket/outfile.zip LIST_OF_S3_FILES

or for local files

    s3packer outfile.zip LIST_OF_S3_FILES

Where `LIST_OF_S3_FILES` should be of the form `s3://BUCKET/KEY`

### License

MIT
