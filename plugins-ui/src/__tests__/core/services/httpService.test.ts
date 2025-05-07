import { get, post, put, remove } from "@/modules/core/services/httpService";
import { describe, it, expect, vi } from "vitest";

const fetchSpy = vi.spyOn(global, "fetch");
const mockFetchSuccessReturn = vi.fn((result: unknown) =>
  Promise.resolve({
    json: () => Promise.resolve(result),
    ok: true,
  } as Response)
);

const mockFetchErrorReturn = vi.fn((result: unknown) =>
  Promise.resolve({
    json: () => Promise.resolve(result),
    ok: false,
  } as Response)
);

const testCases = [
  [
    "get",
    // Note: get function does not accept body parameter so we have to move additionalOptions
    {
      headers: {
        "Method-Type": "GET",
      },
    },
    {
      headers: {
        "Content-Type": "application/json",
        "Method-Type": "GET",
      },
      method: "GET",
    },
    undefined,
    get,
  ],
  [
    "post",
    { test: "test" },
    {
      body: '{"test":"test"}',
      headers: {
        "Method-Type": "POST",
      },
      method: "POST",
    },
    {
      headers: {
        "Method-Type": "POST",
      },
    },
    post,
  ],
  [
    "put",
    { test: "test" },
    {
      body: '{"test":"test"}',
      headers: {
        "Content-Type": "application/json",
        "Method-Type": "PUT",
      },
      method: "PUT",
    },
    {
      headers: {
        "Method-Type": "PUT",
      },
    },
    put,
  ],
  [
    "remove",
    { test: "test" },
    {
      body: '{"test":"test"}',
      headers: {
        "Content-Type": "application/json",
        "Method-Type": "DELETE",
      },
      method: "DELETE",
    },
    {
      headers: {
        "Method-Type": "DELETE",
      },
    },
    remove,
  ],
];

describe("httpService", () => {
  describe.for(testCases)(
    "%s",
    ([, body, options, additionalOptions, func]) => {
      const typedFunc = func as (
        path: string,
        body?: unknown,
        additionalOptions?: unknown
      ) => Promise<Response>;

      it("should call fetch with correct options", async () => {
        fetchSpy.mockReturnValue(mockFetchSuccessReturn({}));
        await typedFunc("test", body, additionalOptions);
        expect(fetchSpy).toBeCalledWith("test", options);
      });
      it("should return parsed response", async () => {
        const mockResponse = { test: "success" };
        fetchSpy.mockReturnValueOnce(mockFetchSuccessReturn(mockResponse));
        const result = await typedFunc("test", body, additionalOptions);
        expect(result).toStrictEqual(mockResponse);
      });
      it("should throw error and return response", async () => {
        const mockResponse = { error: "Test error" };
        fetchSpy.mockReturnValueOnce(mockFetchErrorReturn(mockResponse));
        await expect(
          typedFunc("test", body, additionalOptions)
        ).rejects.toThrowError(mockResponse.error);
      });
      it("should throw error and log in console", async () => {
        fetchSpy.mockImplementationOnce(() => Promise.reject("Test error"));
        const consoleSpy = vi.spyOn(console, "error");
        await expect(
          typedFunc("test", body, additionalOptions)
        ).rejects.toThrowError("Test error");
        expect(consoleSpy).toBeCalledWith("HTTP Error:", "Test error");
      });
    }
  );
});
