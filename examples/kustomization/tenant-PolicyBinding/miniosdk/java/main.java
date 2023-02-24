import io.minio.ListObjectsArgs;
import io.minio.MinioClient;
import io.minio.Result;
import io.minio.errors.MinioException;
import io.minio.messages.Item;
import java.io.IOException;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;

public class ListObjects {
    public static void main(String[] args) throws Exception{
        try {
            String operatorEndpoint = "http://operator.minio-operator.svc.cluster.local:4221/sts/";
            String minioEndpoint = "http://minio.minio-tenant-1.svc.cluster.local";
            SSLSocketFactory sslSocketFactory = null;
            X509TrustManager trustManager = null;

            Provider provider = new CertificateIdentityProvider(operatorEndpoint, sslSocketFactory, trustManager, null,
                    null);

            /* play.min.io for test and development. */
            MinioClient minioClient = MinioClient.builder()
                    .endpoint("https://MINIO-HOST:MINIO-PORT")
                    .credentialsProvider(provider)
                    .build();

            // Lists objects information.
            Iterable<Result<Item>> results = minioClient.listObjects(ListObjectsArgs.builder().bucket("kafka").build());

            for (Result<Item> result : results) {
                Item item = result.get();
                System.out.println(item.lastModified() + "\t" + item.size() + "\t" + item.objectName());
            }
        } catch (MinioException e) {
            System.out.println("Error occurred: " + e);
        }
    }
}