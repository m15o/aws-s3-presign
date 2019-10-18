# aws-s3-presign

Amazon S3のpre-signed URLを生成する。このpre-signed URLを知っていれば誰でもそのS3オブジェクトにアクセスできる。

## 何をするものか

* 基本的にはawscliの`aws s3 presign`と同じ
  <https://docs.aws.amazon.com/cli/latest/reference/s3/presign.html>
* 任意のメソッドを許可することができるので、URLを渡してPUTさせることができる

## 使い方

### bucket-nameのkey.zipをダウンロードできるURLを生成する

`aws-s3-presign https://s3-us-west-2.amazonaws.com/bucket-name/key.zip`

### bucket-nameのkey.zipをファイルをアップロードできるURLを生成する

`aws-s3-presign -method PUT s3://bucket-name/key.zip`

### curlでアクセスする

```bash
SIGNED_URL=$(aws-s3-presign -method PUT s3://bucket-name/key.zip)
curl -f -X PUT "$SIGNED_URL" --data-raw @/path/to/file
```

## 補足

presignしたユーザーにs3のアクセス権がなければpresignしたとしても当然無効になる
