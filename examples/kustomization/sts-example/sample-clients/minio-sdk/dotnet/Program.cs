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
using Minio.Exceptions;
using Minio.DataModel;
using System.Threading.Tasks;
using System.IO;

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

            string caFile;

            if (FileExists(stsCAPath))
            {
                caFile = stsCAPath;
            }
            else
            {
                caFile = kubeRootCAPath;
            }

            try {
                HttpClient client = new HttpClient();
                client.

                Minio.Credentials.ClientProvider credentialsProvider = new Minio.Credentials.WebIdentityProvider();
                var credentials = credentialsProvider.GetCredentials();
                System.Console.WriteLine($"AccessKey: ${credentials.AccessKey}");
                System.Console.WriteLine($"AccessKey: ${credentials.SecretKey}");
                System.Console.WriteLine($"AccessKey: ${credentials.SessionToken}");

                var minio = new MinioClient().WithCredentialsProvider(credentialsProvider)
                                    .WithEndpoint(tenantEndpoint)
                                    .WithHttpClient()
                                    .WithSSL()
                                    .Build();
                FileUpload.Run(minio).Wait();
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
            }
            Console.ReadLine();


        }


        private byte[] GetFile(string? path)
        {
            if (!FileExists(path))
            {
                throw new Exception($"File {path} not found");
            }




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
