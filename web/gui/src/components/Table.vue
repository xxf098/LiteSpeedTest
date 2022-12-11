<template>
  <el-table-v2
    :columns="columns"
    :data="result"
    :width="1700"
    :height="800"
    fixed
  />
</template>

<script>
export default {
  props: ['result']
}
</script>

<script setup>

let id = 0

const columns = [
  {
    key: "remark",
    title: 'Remark',
    dataKey: 'remark',
    width: 800,
    align: "center",
  },
  {
    key: "server",
    title: 'Server',
    dataKey: 'server',
    width: 300,
    align: "center",
  },
  {
    key: "protocol",
    title: 'Protocol',
    dataKey: 'protocol',
    width: 120,
    align: "center",
  },
  {
    key: "ping",
    title: 'Ping',
    dataKey: 'ping',
    width: 100,
    align: "center",
  },
  {
    key: "speed",
    title: 'AvgSpeed',
    dataKey: 'speed',
    width: 180,
    align: "center",
  },
  {
    key: "maxspeed",
    title: 'MaxSpeed',
    dataKey: 'maxspeed',
    width: 180,
    align: "center",
  },
]

</script>


