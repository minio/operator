import { useState } from "react";
import api from "../../../../common/api";
import { ErrorResponseHandler } from "../../../../common/types";

type NoReturnFunction = (param?: any) => void;
type ApiMethodToInvoke = (method: string, url: string, data?: any) => void;
type IsApiInProgress = boolean;

const useApi = (
  onSuccess: NoReturnFunction,
  onError: NoReturnFunction,
): [IsApiInProgress, ApiMethodToInvoke] => {
  const [isLoading, setIsLoading] = useState<boolean>(false);

  const callApi = (method: string, url: string, data?: any, headers?: any) => {
    setIsLoading(true);
    api
      .invoke(method, url, data, headers)
      .then((res: any) => {
        setIsLoading(false);
        onSuccess(res);
      })
      .catch((err: ErrorResponseHandler) => {
        setIsLoading(false);
        onError(err);
      });
  };

  return [isLoading, callApi];
};

export default useApi;
