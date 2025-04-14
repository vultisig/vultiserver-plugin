/**
 * Handles HTTP responses by checking the status and returning the parsed JSON.
 * Throws an error if the response is not ok.
 */
const handleResponse = async (response: Response) => {
  if (!response.ok) {
    const errorData = await response.json();

    throw new Error(
      errorData.error || errorData.message || "Something went wrong"
    );
  }

  // Parse and return the response JSON
  try {
    return await response.json();
  } catch {
    return;
  }
};

/**
 * Handles HTTP errors in a consistent way.
 */
const handleError = (error: unknown) => {
  console.error("HTTP Error:", error);
  throw error;
};

/**
 * Simple helper to merge RequestInit objects
 * @param defaultOptions
 * @param additionalOptions
 * @returns {RequestInit}
 */
const mergeRequestOptions = (
  defaultOptions: RequestInit,
  additionalOptions?: RequestInit
) => {
  if (!additionalOptions) return defaultOptions;
  const combinedHeaders = Object.assign(
    {},
    defaultOptions.headers,
    additionalOptions.headers
  );
  delete defaultOptions["headers"];
  delete additionalOptions["headers"];
  return Object.assign({}, defaultOptions, additionalOptions, {
    headers: { ...combinedHeaders },
  });
};

/**
 * Performs a POST request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - The data to send in the body of the request.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const post = async (
  endpoint: string,
  data: unknown,
  options?: RequestInit
) => {
  try {
    const response = await fetch(endpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...options?.headers,
      },
      body: JSON.stringify(data),
      ...options,
    });
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a GET request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const get = async (endpoint: string, options?: RequestInit) => {
  try {
    const response = await fetch(
      endpoint,
      mergeRequestOptions(
        {
          method: "GET",
          headers: {
            "Content-Type": "application/json",
          },
        },
        options
      )
    );
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a PUT request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - The data to send in the body of the request.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const put = async (
  endpoint: string,
  data: unknown,
  options?: RequestInit
) => {
  try {
    const response = await fetch(
      endpoint,
      mergeRequestOptions(
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        },
        options
      )
    );
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};

/**
 * Performs a DELETE request.
 * @param {string} endpoint - The API endpoint.
 * @param {Object} data - Signature for the policy deletion.
 * @param {Object} options - Additional fetch options (e.g., headers).
 */
export const remove = async (
  endpoint: string,
  data: unknown,
  options?: RequestInit
) => {
  try {
    const response = await fetch(
      endpoint,
      mergeRequestOptions(
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(data),
        },
        options
      )
    );
    return handleResponse(response);
  } catch (error) {
    handleError(error);
  }
};
