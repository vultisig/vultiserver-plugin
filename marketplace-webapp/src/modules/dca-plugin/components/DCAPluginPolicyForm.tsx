import { FormProvider, useForm } from "react-hook-form";
import "./DCAPolicy.css";
import swapIcon from "@/assets/Swap.svg";
import usdcIcon from "@/assets/USDC.png";
import wethIcon from "@/assets/WETH.png";
import { allocate_from_validation, orders_validation, time_period_validation } from "@/modules/dca-plugin/utils/inputSpecifications";
import ToggleSwitch from "@/modules/core/components/ui/toggle-switch/ToggleSwitch";
import SelectBox from "@/modules/core/components/ui/select-box/SelectBox";
import { Input } from "@/modules/core/components/ui/input/Input";
import { v4 as uuidv4 } from 'uuid';
import DCAService from "../services/dcaService";
import { Policy } from "../models/policy";


const DCAPluginPolicyForm = () => {
    const methods = useForm();

    const onSubmit = methods.handleSubmit(async data => {
        const policy: Policy = {
            id: uuidv4(), // todo move to BE
            public_key: "8540b779a209ef961bf20618b8e22c678e7bfbad37ec0",
            plugin_type: "dca",
            policy: {
                chain_id: "1", // hardcoded for now
                source_token_id: "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2", // WETH
                destination_token_id: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
                total_amount: data.allocateAmount,
                total_orders: data.orders,
                schedule: {
                    frequency: data.every,
                    start_time: ""
                }
            },
        }

        try {
            await DCAService.createPolicy(policy);
        } catch (error: any) {
            console.error('Failed to create policy:', error.message);
        }
    })



    return (
        <FormProvider {...methods}>
            <form
                onSubmit={e => e.preventDefault()}
                noValidate
                autoComplete="off"
            >
                <div className="input-field-inline">
                    <div>
                        <Input {...allocate_from_validation} />
                        {/* todo do not hardcode */}
                        <div className="dollar-equivalent">$ 119</div>
                    </div>
                    {/* todo at some point this will no longer be needed */}
                    <div className="display-flex">
                        <img src={usdcIcon} alt="" width="24px" height="24px" />
                        <div>&nbsp;USDC</div>
                    </div>
                </div>
                <button className="swap-btn">
                    <img src={swapIcon} alt="" />
                </button>
                <div className="input-field-inline" style={{ flexDirection: "column", alignItems: "flex-start", color: "#FFFFFF" }}>
                    <div>
                        To Buy
                    </div>
                    {/* todo at some point this will no longer be needed */}
                    <div className="display-flex">
                        <img src={wethIcon} alt="" width="24px" height="24px" />
                        <div>&nbsp;WETH</div>
                    </div>
                </div>
                <div className="display-flex">

                    <div className="input-field-outline">

                        <div className="input-container">
                            <Input {...time_period_validation} />
                            <SelectBox />
                        </div>
                    </div>

                    <div className="input-field-outline">
                        <div className="input-container">
                            <Input {...orders_validation} />
                            <div className="absolute">orders</div>
                        </div>
                    </div>
                </div>
                <div className="display-flex white-text m-t-b-24">
                    <div>Enable given policy</div>
                    <ToggleSwitch />
                </div>
                <button
                    onClick={onSubmit}
                    className="submit"
                >
                    Start
                </button>
            </form>
        </FormProvider>
    );
};

export default DCAPluginPolicyForm;