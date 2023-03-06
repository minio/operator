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

using System;
using Minio;
using System.Threading.Tasks;
using System.Net.Http;
using System.Security.Cryptography.X509Certificates;

namespace sts
{
    class Example
    {
        static void Main(string[] args)
        {
            var tenantEndpoint = Environment.GetEnvironmentVariable("MINIO_ENDPOINT");
            var stsEndpoint = Environment.GetEnvironmentVariable("STS_ENDPOINT");
            var tenantNamespace = Environment.GetEnvironmentVariable("TENANT_NAMESPACE");
            var bucketName = Environment.GetEnvironmentVariable("BUCKET");
            var kubeRootCAPath = Environment.GetEnvironmentVariable("KUBERNETES_CA_PATH");
            var stsCAPath = Environment.GetEnvironmentVariable("STS_CA_PATH");

            Environment.SetEnvironmentVariable("AWS_ROLE_ARN","arn:aws:iam::111111111:dummyroot");
            Environment.SetEnvironmentVariable("AWS_ROLE_SESSION_NAME","optional-session-name");

            string? caFile = "";

            if (FileExists(stsCAPath))
            {
                caFile = stsCAPath;
            }
            else
            {
                if (FileExists(kubeRootCAPath))
                {
                    caFile = kubeRootCAPath;
                }
            }

            try
            {
                var tenantEndpointUrl = new Uri(tenantEndpoint);
                var credentialsProvider = new Minio.Credentials.IAMAWSProvider();
                using var minioClient = new MinioClient()
                    .WithEndpoint(tenantEndpointUrl.Host, tenantEndpointUrl.Port)
                    .WithSSL()
                    .WithCredentialsProvider(credentialsProvider)
                    .WithHttpClient(GetHttpTransport(caFile))
                    .Build();

                var url = new Uri($"{stsEndpoint}/{tenantNamespace}");
                credentialsProvider = credentialsProvider
                    .WithEndpoint(url.ToString)
                    .WithMinioClient(minioClient);

                credentialsProvider.Validate();
                 
                var credentials = credentialsProvider.GetCredentials();
                System.Console.WriteLine($"AccessKey: ${credentials.AccessKey}");
                System.Console.WriteLine($"AccessKey: ${credentials.SecretKey}");
                System.Console.WriteLine($"AccessKey: ${credentials.SessionToken}");

                ListBuckets(minioClient).GetAwaiter().GetResult();
                ListObjects(minioClient, bucketName).GetAwaiter().GetResult();
            }
            catch (UriFormatException uer)
            {
                Console.WriteLine($"STS endpoint malformed: {uer.Message}");
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
                Console.WriteLine(ex.StackTrace);
                Environment.Exit(111);
            }
        }

        public static async Task ListBuckets(IMinioClient minio)
        {
            try
            {
                Console.WriteLine("Running example for API: ListBucketsAsync");
                var list = await minio.ListBucketsAsync().ConfigureAwait(false);
                foreach (var bucket in list.Buckets) Console.WriteLine($"{bucket.Name} {bucket.CreationDateDateTime}");
                Console.WriteLine();
            }
            catch (Exception e)
            {
                Console.WriteLine($"[Bucket]  Exception: {e}");
            }
        }

        public static async Task ListObjects(IMinioClient minio, string bucketName)
        {
            try
            {
                var listArgs = new ListObjectsArgs()
                    .WithBucket(bucketName)
                    .WithRecursive(true);
                var observable = minio.ListObjectsAsync(listArgs);
                var subscription = observable.Subscribe(
                    item => Console.WriteLine($"Object: {item.Key}"),
                    ex => Console.WriteLine($"OnError: {ex}"),
                    () => Console.WriteLine($"Listed all objects in bucket {bucketName}\n"));
            }
            catch (System.Exception e)
            {
                Console.WriteLine($"[Object]  Exception: {e}");
            }
        }

        private static HttpClient GetHttpTransport(string caPath)
        {
            var handler = new HttpClientHandler();
            if (!string.IsNullOrEmpty(caPath))
            {
                handler.ServerCertificateCustomValidationCallback = (message, cert, chain, _) =>
                {
                    chain.ChainPolicy.TrustMode = X509ChainTrustMode.CustomRootTrust;
                    chain.ChainPolicy.CustomTrustStore.Add(new X509Certificate2(caPath));
                    return chain.Build(cert);
                };
            }

            var httpClient = new HttpClient(handler);
            return httpClient;
        }

        private static bool FileExists(string? path)
        {
            if (String.IsNullOrEmpty(path))
            {
                return false;
            }
            return false;
        }
    }
}
