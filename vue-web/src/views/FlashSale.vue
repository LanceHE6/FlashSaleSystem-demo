<template>
    <div class="cover-container">
        <n-card class="cover">
            <n-space vertical>
                <n-button type="info" @click="addGoods">
                添加商品
                </n-button>
                <n-modal
                        v-model:show="addModalVisible"
                        preset="dialog"
                        title="添加商品"
                        positive-text="提交"
                        negative-text="取消"
                        @positive-click="submitAddGoods"
                        @negative-click="cancelCallback"
                    >
                        <n-form :model="addData">
                            <n-form-item label="商品名称">
                                <n-input v-model:value="addData.name" type= "text"></n-input>
                            </n-form-item>
                            <n-form-item label="数量">
                                <n-input v-model:value="addData.quantity"></n-input>
                            </n-form-item>
                        </n-form>

                    </n-modal>

                <n-card 
                    v-for="(item, index) in items" 
                    :key="item.gid" 
                    :title="'#' + (index+1) + '.' + item.name" 
                    size="large"
                >
                    <h5>商品编码：{{ item.gid }} 商品更新时间：{{ item.updated_at }}</h5>
                    <h3>数量：{{ item.quantity }}</h3>
                    
                    <n-space>
                        <n-button type="primary" @click="showModal(item.gid)">秒杀</n-button>
                        <n-button type="info" @click="updateGoods(item.gid, item.name, item.quantity)">更新</n-button>
                        <n-button type="error" @click="delGoods(item.gid)">删除</n-button>
                    </n-space>
                    
                    <n-modal
                        v-model:show="modalVisible"
                        preset="dialog"
                        title="填写订单信息"
                        positive-text="提交"
                        negative-text="取消"
                        @positive-click="submitOrder"
                        @negative-click="cancelCallback"
                    >
                        <n-form :model="order" ref="formRef">
                            <n-form-item label="商品ID">
                                <n-input v-model:value="order.goods_id" type= "text" :disabled = "true" ></n-input>
                            </n-form-item>
                            <n-form-item path="user_id" label="用户ID">
                                <n-input v-model:value="order.user_id"></n-input>
                            </n-form-item>
                            <n-form-item path="order_num" label="下单数量">
                                <n-input v-model:value="order.order_num"></n-input>
                            </n-form-item>
                        </n-form>

                    </n-modal>

                    <n-modal
                        v-model:show="updateModalVisible"
                        preset="dialog"
                        title="更新 {{item.name}} 数据"
                        positive-text="提交"
                        negative-text="取消"
                        @positive-click="submitUpdate"
                        @negative-click="cancelCallback"
                    >
                        <n-form :model="updateData">
                            <n-form-item label="商品ID">
                                <n-input v-model:value="updateData.goods_id" type= "text" :disabled = "true" ></n-input>
                            </n-form-item>
                            <n-form-item  label="商品名称">
                                <n-input v-model:value="updateData.name"></n-input>
                            </n-form-item>
                            <n-form-item  label="商品数量">
                                <n-input v-model:value="updateData.quantity"></n-input>
                            </n-form-item>
                        </n-form>

                    </n-modal>

                    <n-modal
                        v-model:show="delModalVisible"
                        preset="dialog"
                        title="确认删除"
                        positive-text="确定"
                        negative-text="取消"
                        @positive-click="submitDel"
                        @negative-click="cancelCallback"
                    >
                    </n-modal>

                </n-card>
            </n-space>
        </n-card>
    </div>
</template>

<script>
import { inject, ref, onMounted, reactive, watchEffect } from 'vue';
import { useMessage } from 'naive-ui';

export default {
    setup() {
        const message = useMessage()
        const axios = inject("axios")

        const items = ref([])

        // 创建 WebSocket 连接
        const socket = new WebSocket('ws://127.0.0.1:8080/api/ws')

        // WebSocket 连接打开时触发
        socket.onopen = function(event) {
            console.log('WebSocket is open now.')
        }

        // WebSocket 连接关闭时触发
        socket.onclose = function(event) {
            console.log('WebSocket is closed now.')
        }

        // WebSocket 接收到消息时触发
        socket.onmessage = function(event) {
            let stockInfo = JSON.parse(event.data)
            // 更新商品的库存信息
            for (let item of items.value) {
                if (stockInfo.hasOwnProperty(item.gid)) {
                    item.quantity = stockInfo[item.gid]
                }
            }
        }

        // WebSocket 出错时触发
        socket.onerror = function(error) {
            console.log(`WebSocket error: ${error}`)
        }

        // 控制添加商品弹窗可见性
        const addModalVisible = ref(false)
        const addData = reactive({
            name: "",
            quantity: "0"
        })
        const addGoods = () => {
            addModalVisible.value = true
        }
        const submitAddGoods = async() => {
            addData.quantity = parseInt(addData.quantity)
            const response = await axios.post("/api/goods/add", addData)
            if (response.status !== 200){
                message.error("error!");
            } else{
                message.success("添加成功")
                loadGoodsList()  
            }
        }

        // 控制修改商品弹窗可见性
        const updateModalVisible = ref(false)
        const updateData = reactive({
            goods_id: "",
            name: "",
            quantity: ""
        })
        const updateGoods = (gid, name, quantity) => {
            updateData.goods_id = gid
            updateData.name = name
            updateData.quantity = quantity
            updateModalVisible.value = true
        }
        const submitUpdate = async() => {
            updateData.quantity = parseInt(updateData.quantity)
            const response = await axios.put("/api/goods/update", updateData)
            if (response.status !== 200){
                message.error("error!");
            } else{
                message.success("更新成功")
                loadGoodsList()  
            }
        }

        // 控制删除商品弹窗可见性
        const delModalVisible = ref(false)
        const delData = reactive({
            goods_id: ""
        })
        const delGoods = (gid) => {
            delData.goods_id = gid
            delModalVisible.value = true
        }
        const submitDel = async() => {
            const response = await axios.post("/api/goods/del", delData)
            if (response.status !== 200){
                message.error("error!");
            } else{
                message.success("删除成功")
                loadGoodsList()  
            }
        }

        // 控制订单弹窗可见性
        const modalVisible = ref(false)
        const order = reactive({
            user_id: "g22122",
            goods_id: "",
            order_num: "1",
            order_time: "",
        })

        const showModal = (gid) =>{
            order.goods_id = gid
            modalVisible.value = true
        }

        
        const cancelCallback = () => {
            
        }

        const submitOrder = async () => {
            order.order_num = parseInt(order.order_num)
            order.order_time = toString(parseInt(new Date().getTime()/1000))
            const response = await axios.post('/api/seckill', order);
            if (response.status !== 200){
                message.error("error!");
            } else{
                message.success("success")  
            }
            
        }

        // 加载数据
        const loadGoodsList = async () => {
            try {
                const response = await axios.get('/api/goods/list');
                items.value = response.data.data.rows;
            } catch (error) {
                console.error(error);
            }
        }

        onMounted(loadGoodsList)

        // 当组件卸载时，关闭 WebSocket 连接
        watchEffect((onInvalidate) => {
            onInvalidate(() => {
                socket.close()
            })
        })

        return {
            modalVisible,
            showModal,
            items,
            order,
            cancelCallback,
            submitOrder,

            addData,
            addModalVisible,
            addGoods,
            submitAddGoods,

            updateModalVisible,
            updateData,
            updateGoods,
            submitUpdate,

            delData,
            delGoods,
            delModalVisible,
            submitDel,
        }
    },
}
</script>


<style scoped>
.cover-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh; /* Optional: This will make the container take up the full height of the viewport */
}

.cover {
    max-width: 600px;
}
</style>
