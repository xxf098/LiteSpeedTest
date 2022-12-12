<template>
  <el-table-v2
    :sort-state="sortState"
    :columns="columns"
    :data="result"
    :width="1700"
    :height="800"
    @column-sort="onSort"
  />
</template>

<script>
export default {
  props: ['result'],
}
</script>

<script setup>
import { ref } from 'vue'
import {
  TableV2SortOrder
} from 'element-plus'
import { Filter } from '@element-plus/icons-vue'


const columns = [
  {
    key: "remark",
    title: 'Remark',
    dataKey: 'remark',
    width: 700,
    align: "center",
    sortable: true,
  },
  {
    key: "server",
    title: 'Server',
    dataKey: 'server',
    width: 400,
    align: "center",
    sortable: true,
  },
  {
    key: "protocol",
    title: 'Protocol',
    dataKey: 'protocol',
    width: 120,
    align: "center",
    sortable: true,
  },
  {
    key: "ping",
    title: 'Ping',
    dataKey: 'ping',
    width: 100,
    align: "center",
    sortable: true,
  },
  {
    key: "speed",
    title: 'AvgSpeed',
    dataKey: 'speed',
    width: 180,
    align: "center",
    sortable: true,
  },
  {
    key: "maxspeed",
    title: 'MaxSpeed',
    dataKey: 'maxspeed',
    width: 180,
    align: "center",
    sortable: true,
  },
]

const sortState = ref({
  'remark': TableV2SortOrder.ASC,
  'server': TableV2SortOrder.ASC,
  'protocol': TableV2SortOrder.ASC,
  'ping': TableV2SortOrder.ASC,
  'speed': TableV2SortOrder.ASC,
  'maxspeed': TableV2SortOrder.ASC,
})

const onSort = ({ key, order }) => {
  sortState.value[key] = order
}

</script>


