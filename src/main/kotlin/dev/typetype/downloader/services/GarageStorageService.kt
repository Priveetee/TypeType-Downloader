package dev.typetype.downloader.services

import dev.typetype.downloader.config.AppConfig
import software.amazon.awssdk.auth.credentials.AwsBasicCredentials
import software.amazon.awssdk.auth.credentials.StaticCredentialsProvider
import software.amazon.awssdk.core.sync.RequestBody
import software.amazon.awssdk.regions.Region
import software.amazon.awssdk.services.s3.S3Client
import software.amazon.awssdk.services.s3.S3Configuration
import software.amazon.awssdk.services.s3.model.CreateBucketRequest
import software.amazon.awssdk.services.s3.model.GetObjectRequest
import software.amazon.awssdk.services.s3.model.HeadBucketRequest
import software.amazon.awssdk.services.s3.model.PutObjectRequest
import software.amazon.awssdk.services.s3.model.S3Exception
import software.amazon.awssdk.services.s3.model.DeleteObjectRequest
import software.amazon.awssdk.services.s3.presigner.S3Presigner
import software.amazon.awssdk.services.s3.presigner.model.GetObjectPresignRequest
import java.net.URI
import java.nio.file.Path
import java.time.Duration

class GarageStorageService(config: AppConfig) {
    private val endpoint = URI(config.s3Endpoint)
    private val region = Region.of(config.s3Region)
    private val bucket = config.s3Bucket
    private val credentials = StaticCredentialsProvider.create(
        AwsBasicCredentials.create(config.s3AccessKey, config.s3SecretKey)
    )

    private val s3: S3Client = S3Client.builder()
        .endpointOverride(endpoint)
        .region(region)
        .credentialsProvider(credentials)
        .serviceConfiguration(S3Configuration.builder().pathStyleAccessEnabled(true).build())
        .build()

    private val presigner: S3Presigner = S3Presigner.builder()
        .endpointOverride(endpoint)
        .region(region)
        .credentialsProvider(credentials)
        .serviceConfiguration(S3Configuration.builder().pathStyleAccessEnabled(true).build())
        .build()

    fun ensureBucket() {
        try {
            s3.headBucket(HeadBucketRequest.builder().bucket(bucket).build())
        } catch (error: S3Exception) {
            if (error.statusCode() == 404) {
                s3.createBucket(CreateBucketRequest.builder().bucket(bucket).build())
            } else {
                throw error
            }
        }
    }

    fun putBytes(objectKey: String, body: ByteArray, contentType: String) {
        val request = PutObjectRequest.builder().bucket(bucket).key(objectKey).contentType(contentType).build()
        s3.putObject(request, RequestBody.fromBytes(body))
    }

    fun putFile(objectKey: String, filePath: Path, contentType: String) {
        val request = PutObjectRequest.builder().bucket(bucket).key(objectKey).contentType(contentType).build()
        s3.putObject(request, RequestBody.fromFile(filePath))
    }

    fun presignGet(objectKey: String, duration: Duration): String {
        val getRequest = GetObjectRequest.builder().bucket(bucket).key(objectKey).build()
        val presignRequest = GetObjectPresignRequest.builder().signatureDuration(duration).getObjectRequest(getRequest).build()
        return presigner.presignGetObject(presignRequest).url().toString()
    }

    fun deleteObject(objectKey: String) {
        val request = DeleteObjectRequest.builder().bucket(bucket).key(objectKey).build()
        s3.deleteObject(request)
    }

    fun close() {
        presigner.close()
        s3.close()
    }
}
