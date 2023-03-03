// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package sts;
import io.minio.ListObjectsArgs;
import io.minio.MinioClient;
import io.minio.Result;
import io.minio.errors.MinioException;
import io.minio.messages.Item;
import io.minio.credentials.CertificateIdentityProvider;
import io.minio.credentials.Provider;
import java.io.IOException;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;

import javax.net.ssl.SSLSocketFactory;
import javax.net.ssl.X509TrustManager;


public class OperatorSTSExample {
    public static void main(String[] args) throws Exception{
        try {
            String operatorEndpoint = System.getenv("OPERATOR_ENDPOINT");
            String minioEndpoint = System.getenv("TENANT_ENDPOINT");
            String tenantNamespace = System.getenv("TENANT_NAMESPACE");
	        String bucketName = System.getenv("BUCKET");

            SSLSocketFactory sslSocketFactory = null;
            X509TrustManager trustManager = null;

            Provider provider = new CertificateIdentityProvider(operatorEndpoint, sslSocketFactory, trustManager, null, null);

            MinioClient minioClient = MinioClient.builder()
                    .endpoint(minioEndpoint)
                    .credentialsProvider(provider)
                    .build();

            // Lists objects information.
            Iterable<Result<Item>> results = minioClient.listObjects(ListObjectsArgs.builder().bucket(bucketName).build());

            for (Result<Item> result : results) {
                Item item = result.get();
                System.out.println(item.lastModified() + "\t" + item.size() + "\t" + item.objectName());
            }
        } catch (MinioException e) {
            System.out.println("Error occurred: " + e);
        }
    }
}