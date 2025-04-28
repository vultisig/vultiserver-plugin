import Toast from "@/modules/core/components/ui/toast/Toast";

type CustomEvents = Toast;
type CustomEventsNames = "onToast";
export const subscribe = (
  eventName: string,
  listener: (event: CustomEventInit<CustomEvents>) => void
) => {
  document.addEventListener(eventName, listener);
};

export const unsubscribe = (
  eventName: string,
  listener: (event: CustomEventInit<CustomEvents>) => void
) => {
  document.removeEventListener(eventName, listener);
};

export const publish = (eventName: CustomEventsNames, data: CustomEvents) => {
  const event = new CustomEvent(eventName, { detail: data });
  document.dispatchEvent(event);
};
