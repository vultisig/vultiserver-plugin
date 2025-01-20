export default function SelectBox() {
    // todo do not hardcode the options
    return (
        <>
            <select aria-label="every" name="selectedTime" defaultValue="minute">
                <option value="minute">minute</option>
                <option value="hour">hour</option>
                <option value="day">day</option>
                <option value="week">week</option>
                <option value="month">month</option>
            </select>
        </>
    );
}